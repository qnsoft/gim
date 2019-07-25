/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         handler.go
@ Create Time:  2019/7/24 22:09
@ Software:     GoLand
*/

package services

import (
	"fmt"
	"gim/src/models"
	"github.com/gomodule/redigo/redis"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// 客户端
type Client struct {
	AppKey string // 认证标识
	Id     string // 客户端唯一ID, 由客户端维护该字段的唯一性
	Name   string // 客户端名称
	City   string // 城市
	Addr   string // 客户端地址
}

// 基础数据结构
type Base struct {
	ServiceName string //服务名称
}

// 聊天室模式
type ChatRoom struct {
	Base
}

// 消息推送模式
type MessagePush struct {
	Base
}

// 消息格式化: 公共广播
func makePublicMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("%s||[%s:%s] -> %s\n", client.AppKey, client.Id, client.Name, msg)
	return
}

// 在线列表: 增加
func (b Base) addOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SADD", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"), id); err != nil {
		log.Println("SADD failed!", err)
	}
}

// 在线列表: 移除
func (b Base) delOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SREM", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"), id); err != nil {
		log.Println("SREM failed!", err)
	}
}

// 在线列表: 获取集合数据
func (b Base) GetOnlineMap(appKey string) ([]string, error) {
	c := models.Pool.Get()
	defer c.Close()
	_ = c.Send("SMEMBERS", strings.Join([]string{appKey, b.ServiceName, "onlineMap"}, ":"))
	_ = c.Flush()
	if reply, err := redis.Strings(c.Receive()); err != nil {
		log.Println("SMEMBERS failed!", err)
		return nil, err
	} else {
		return reply, nil
	}
}

// 频道发布
func (b Base) Publish(channel, msg string, public bool) {
	c := models.Pool.Get()
	defer c.Close()
	switch public {
	// 发给公共频道 -> public:Broadcast
	case true:
		if _, err := c.Do("PUBLISH", strings.Join([]string{b.ServiceName, "public", channel}, ":"), msg); err != nil {
			log.Println("PUBLISH failed", err)
		}
	// 发给私有频道 -> appkey:unique
	case false:
		if _, err := c.Do("PUBLISH", channel, msg); err != nil {
			log.Println("PUBLISH failed", err)
		}
	}
}

// 频道订阅(特指私人频道)
func (b Base) Subscribe(channel string, conn net.Conn) {
	c := models.Pool.Get()
	psc := redis.PubSubConn{Conn: c}
	_ = psc.Subscribe(strings.Join([]string{b.ServiceName, channel}, ":"))
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if _, err := conn.Write([]byte(v.Data)); err != nil {
				log.Println("Send data error: ", err)
				return
			}
		case redis.Subscription:
		case error:
			log.Printf("Unknown type: %+v, %T", v, v)
			return
		}
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := strings.Join([]string{client.AppKey, client.Id}, ":")
	c.addOnlineMap(unique, client.AppKey)

	// 广播用户上线
	c.Publish("Broadcast", makePublicMessage(client, "Login"), true)

	// 订阅私人频道
	go c.Subscribe(unique, conn)
	// 发送欢迎语
	if _, err := conn.Write([]byte("Welcome to the GIM ChatRoom mode ^_^")); err != nil {
		log.Println("Send welcome failed!", err)
		return
	}

	// 接收当前用户输入数据
	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := conn.Read(buf); err != nil {
				offline <- true
				if err == io.EOF {
					log.Println("Customer has closed the connection")
					return
				}
				log.Println("Recv data error: ", err)
				return
			} else {
				// 广播用户数据
				c.Publish("Broadcast", makePublicMessage(client, string(buf[:n])), true)
				online <- true
			}
		}
	}()

	// 连接超时退出
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			c.delOnlineMap(unique, client.AppKey)
			c.Publish("Broadcast", makePublicMessage(client, "Logout"), true)
			return
		// 超时退出
		case <-time.After(300 * time.Second):
			c.delOnlineMap(unique, client.AppKey)
			c.Publish("Broadcast", makePublicMessage(client, "Time out"), true)
			return
		}
	}
}

// 基于消息推送的连接处理
func (m MessagePush) HandleConnection(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := strings.Join([]string{client.AppKey, client.Id}, ":")
	m.addOnlineMap(unique, client.AppKey)

	// 订阅私人频道
	go m.Subscribe(unique, conn)
	// 发送欢迎语
	if _, err := conn.Write([]byte("Welcome to the GIM MessagePush mode ^_^")); err != nil {
		log.Println("Send welcome failed!", err)
		return
	}

	// 退出逻辑
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			m.delOnlineMap(unique, client.AppKey)
			return
		}
	}
}
