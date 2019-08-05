/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         handler.go
@ Create Time:  2019-07-29 14:54
@ Software:     GoLand
*/

package services

import (
	"encoding/json"
	"fmt"
	"gim/app/server/models"
	"github.com/gomodule/redigo/redis"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type Client struct {
	AppKey string      `json:"app_key" binding:"required"` // 认证标识
	Token  string      `json:"token" binding:"required"`   // 认证令牌
	Id     string      `json:"id" binding:"required"`      // 客户端唯一ID, 由客户端维护该字段的唯一性
	Name   string      `json:"name" binding:"required"`    // 客户端名称
	City   string      `json:"city"`                       // 城市
	Addr   string      // 客户端地址
	C      chan string // 私有消息频道
}

type PublicMessage struct {
	AppKey   string `json:"app_key"`
	From     string `json:"from"`
	FormName string `json:"form_name"`
	To       string `json:"to"`
	Content  string `json:"content"`
}

// 基础数据结构
type Base struct {
	ServiceName string            //服务名称
	OnlineMap   map[string]Client // 客户端在线列表, e.g. {"id": client object}
}

// 聊天室模式
type ChatRoom struct {
	Base
}

// 消息推送模式
type MessagePush struct {
	Base
}

// 在线列表: 增加
func (b Base) addOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SADD", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"), id); err != nil {
		log.Println("Redis: SADD failed, ", err)
	}
}

// 在线列表: 移除
func (b Base) delOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SREM", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"), id); err != nil {
		log.Println("Redis: SREM failed, ", err)
	}
}

// 在线列表: 获取集合数据
func (b Base) GetOnlineMap(appKey string) ([]string, error) {
	c := models.Pool.Get()
	defer c.Close()
	_ = c.Send("SMEMBERS", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"))
	_ = c.Flush()
	if reply, err := redis.Strings(c.Receive()); err != nil {
		log.Println("Redis: SMEMBERS failed, ", err)
		return nil, err
	} else {
		return reply, nil
	}
}

// 公共频道消息格式化
func PublicMessageBuilder(msg string, c Client) string {
	buf, _ := json.Marshal(PublicMessage{c.AppKey, c.Id, c.Name, "all", msg})
	return string(buf)
}

// 频道发布
func (b Base) Publish(msg string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("PUBLISH", strings.Join([]string{b.ServiceName, "Broadcast"}, ":"), msg); err != nil {
		log.Println("Redis: Channel failed to post message, ", err)
	}
}

// 频道订阅、消息转发
func (b Base) Subscribe() {
	c := models.Pool.Get()
	defer c.Close()
	psc := redis.PubSubConn{Conn: c}
	_ = psc.Subscribe(strings.Join([]string{b.ServiceName, "Broadcast"}, ":"))
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			var msg PublicMessage
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				log.Println("Subscribe: Json Unmarshal failed, ", err)
				continue
			}

			// 消息转发、TODO 定向推送
			for _, obj := range b.OnlineMap {
				if msg.AppKey == obj.AppKey {
					obj.C <- fmt.Sprintf("[%s:%s] -> %s", msg.From, msg.FormName, msg.Content)
				}
			}
		case redis.Subscription:
		case error:
			log.Printf("Subscribe->Receive(): Unknown type, %T, %+v\n", v, v)
			return
		}
	}
}

// 客户端初始化
func InitClient(conn net.Conn) (client Client, err error) {
	buf := make([]byte, 1024)
	for {
		if n, _ := conn.Read(buf); n < 1024 {
			if err = json.Unmarshal(buf[:n], &client); err != nil {
				return client, err
			}
			client.Addr = conn.RemoteAddr().String()
			client.C = make(chan string)
			return client, nil
		}
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	c.OnlineMap[client.Id] = client
	c.addOnlineMap(client.Id, client.AppKey)

	// 监听私人频道
	go func() {
		for msg := range client.C {
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Printf("HandleConnection->%s: Client connection is not available, %s\n", c.ServiceName, err)
				return
			}
		}
	}()

	// 广播用户上线
	c.Publish(PublicMessageBuilder("Login", client))

	// 发送欢迎语
	client.C <- "Welcome to the GIM ChatRoom mode ^_^"

	// 监听输入
	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := conn.Read(buf); err != nil {
				offline <- true
				if err == io.EOF {
					log.Printf("HandleConnection->%s: Client has closed the connection\n", c.ServiceName)
					return
				}
				log.Printf("HandleConnection->%s: Receive input failed\n", err)
				return
			} else {
				msg := string(buf[:n])
				switch msg {
				case "Heartbeat:ack":
				default:
					// 广播用户数据
					c.Publish(PublicMessageBuilder(string(buf[:n]), client))
				}
				online <- true
			}
		}
	}()

	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(c.OnlineMap, client.Id)
			c.delOnlineMap(client.Id, client.AppKey)
			c.Publish(PublicMessageBuilder("Signout", client))
			return
		// 超时退出
		case <-time.After(360 * time.Second):
			delete(c.OnlineMap, client.Id)
			c.delOnlineMap(client.Id, client.AppKey)
			c.Publish(PublicMessageBuilder("Timeout", client))
			return
		}
	}
}

// 基于消息推送的连接处理
func (m MessagePush) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	m.OnlineMap[client.Id] = client
	m.addOnlineMap(client.Id, client.AppKey)

	// 监听私人频道
	go func() {
		for msg := range client.C {
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Printf("HandleConnection->%s: Client connection is not available, %s\n", m.ServiceName, err)
				return
			}
		}
	}()

	// 发送欢迎语
	client.C <- "Welcome to the GIM MessagePush mode ^_^"

	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(m.OnlineMap, client.Id)
			m.delOnlineMap(client.Id, client.AppKey)
			return
		// 超时退出
		case <-time.After(360 * time.Second):
			delete(m.OnlineMap, client.Id)
			m.delOnlineMap(client.Id, client.AppKey)
			return
		}
	}
}
