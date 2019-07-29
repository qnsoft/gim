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

// 定义全局在线列表内结构
var onlineMap = make(map[string]Client)

type Client struct {
	AppKey string      `json:"app_key" binding:"required"` // 认证标识
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
	ServiceName string                       //服务名称
	OnlineMap   map[string]map[string]Client // 客户端在线列表
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
		log.Println("Redis: SADD failed", err)
	}
}

// 在线列表: 移除
func (b Base) delOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SREM", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"), id); err != nil {
		log.Println("Redis: SREM failed", err)
	}
}

// 在线列表: 获取集合数据
func (b Base) GetOnlineMap(appKey string) ([]string, error) {
	c := models.Pool.Get()
	defer c.Close()
	_ = c.Send("SMEMBERS", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"))
	_ = c.Flush()
	if reply, err := redis.Strings(c.Receive()); err != nil {
		log.Println("Redis: SMEMBERS failed", err)
		return nil, err
	} else {
		return reply, nil
	}
}

// 公共频道消息格式化
func PublicMessageBuilder(c Client, msg string) string {
	buf, _ := json.Marshal(PublicMessage{c.AppKey, c.Id, c.Name, "all", msg})
	return string(buf)
}

// 频道发布
func (b Base) Publish(msg string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("PUBLISH", strings.Join([]string{b.ServiceName, "Broadcast"}, ":"), msg); err != nil {
		log.Println("Redis: Channel failed to post message", err)
	}
}

// 频道订阅
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
				log.Println("Json Unmarshal failed", err)
				continue
			}

			// 消息转发
			for _, obj := range b.OnlineMap[msg.AppKey] {
				obj.C <- fmt.Sprintf("[%s:%s] -> %s", msg.From, msg.FormName, msg.Content)
			}
		case redis.Subscription:
		case error:
			log.Printf("Unknown type, %T, %+v\n", v, v)
			return
		}
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	onlineMap[client.Id] = client
	c.OnlineMap[client.AppKey] = onlineMap
	c.addOnlineMap(client.Id, client.AppKey)

	// 监听私人频道
	go func() {
		for msg := range client.C {
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Println("Connection client exception", err)
				return
			}
		}
	}()

	// 广播用户上线
	c.Publish(PublicMessageBuilder(client, "Login"))

	// 发送欢迎语
	client.C <- "Welcome to the GIM ChatRoom mode ^_^"

	// 监听输入
	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := conn.Read(buf); err != nil {
				offline <- true
				if err == io.EOF {
					log.Println("Customer has closed the connection")
					return
				}
				log.Println("Receive input error", err)
				return
			} else {
				// 广播用户数据
				fmt.Println(string(buf[:n]))
				c.Publish(PublicMessageBuilder(client, string(buf[:n])))
				online <- true
			}
		}
	}()

	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(c.OnlineMap[client.AppKey], client.Id)
			c.delOnlineMap(client.Id, client.AppKey)
			c.Publish(PublicMessageBuilder(client, "Signout"))
			return
		// 超时退出
		case <-time.After(360 * time.Second):
			delete(c.OnlineMap[client.AppKey], client.Id)
			c.delOnlineMap(client.Id, client.AppKey)
			c.Publish(PublicMessageBuilder(client, "Timeout"))
			return
		}
	}
}

// 基于消息推送的连接处理
func (m MessagePush) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	onlineMap[client.Id] = client
	m.OnlineMap[client.AppKey] = onlineMap
	m.addOnlineMap(client.Id, client.AppKey)

	// 监听私人频道
	go func() {
		for msg := range client.C {
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Println("Connection client exception", err)
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
			delete(m.OnlineMap[client.AppKey], client.Id)
			m.delOnlineMap(client.Id, client.AppKey)
			return
		// 超时退出
		case <-time.After(360 * time.Second):
			delete(m.OnlineMap[client.AppKey], client.Id)
			m.delOnlineMap(client.Id, client.AppKey)
			return
		}
	}
}
