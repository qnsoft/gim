/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         handler.go
@ Create Time:  2019/7/18 22:56
@ Software:     GoLand
*/

package im

import (
	"fmt"
	"gim/src/models"
	"github.com/gomodule/redigo/redis"
	"io"
	"log"
	"net"
	"time"
)

// 集群模式在线列表: 增加
func (b Base) addOnlineMap(id string) {
	c := models.Pool.Get()
	defer c.Close()
	_, err := c.Do("SADD", "appkey:onlineMap", id)
	if err != nil {
		log.Println("SADD failed!", err)
	}
}

// 集群模式在线列表: 移除
func (b Base) delOnlineMap(id string) {
	c := models.Pool.Get()
	defer c.Close()
	_, err := c.Do("SREM", "appkey:onlineMap", id)
	if err != nil {
		log.Println("SREM failed!", err)
	}
}

// 集群模式在线列表: 获取集合数据
func (b Base) getOnlineMap() {
	c := models.Pool.Get()
	defer c.Close()
	_ = c.Send("SMEMBERS", "appkey:onlineMap")
	_ = c.Flush()
	data, err := c.Receive()
	if err != nil {
		log.Println("SMEMBERS failed!", err)
		return
	}
	switch data.(type) {
	case redis.Message:
		log.Println("Message: ", data)
	default:
		log.Printf("SMEMBERS: %v", data)
	}
}

// 集群模式: 频道发布
func (b Base) Publish(channel, msg string) {
	c := models.Pool.Get()
	defer c.Close()
	_, err := c.Do("PUBLISH", "appkey:"+channel, msg)
	if err != nil {
		log.Println("PUBLISH failed", err)
	}
}

// 集群模式: 频道订阅
func (b Base) Subscribe() {

}

// 基于聊天室的连接处理
func (c ChatRoom) standaloneHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := fmt.Sprintf("%s<->%s", client.Id, client.Addr)
	c.onlineMap[unique] = client
	// 广播用户上线
	c.Broadcast <- makeMessage(client, "Login")

	// 向当前用户发送数据
	go func() {
		for msg := range client.C {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Send data error: ", err)
				return
			}
		}
	}()

	// 向当前用户发送欢迎语
	client.C <- makeMessage(client, "Welcome to the GIM ChatRoom mode ^_^")

	// 接收当前用户输入数据
	go func() {
		buf := make([]byte, 2*1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				offline <- true
				if err == io.EOF {
					log.Println("Customer has closed the connection")
					return
				}
				log.Println("Recv data error: ", err)
				return
			}
			// 广播用户数据
			c.Broadcast <- makeMessage(client, string(buf[:n]))
			online <- true
		}
	}()

	// 连接超时退出
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(c.onlineMap, unique)
			c.Broadcast <- makeMessage(client, "Logout")
			return
		// 超时退出
		case <-time.After(300 * time.Second):
			delete(c.onlineMap, unique)
			c.Broadcast <- makeMessage(client, "Time out")
			return
		}
	}
}

// 基于消息推送的连接处理
func (m MessagePush) standaloneHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := fmt.Sprintf("%s<->%s", client.Id, client.Addr)
	m.onlineMap[unique] = client

	// 向当前用户发送数据
	go func() {
		for msg := range client.C {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Send data error: ", err)
				return
			}
		}
	}()

	// 向当前用户发送欢迎语
	client.C <- makeMessage(client, "Welcome to the GIM MessagePush mode ^_^")

	// 连接超时退出
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(m.onlineMap, unique)
			return
		}
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) clusterHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := fmt.Sprintf("%s<->%s", client.Id, client.Addr)
	c.addOnlineMap(unique)
	// 广播用户上线
	c.Publish("Broadcast", makeMessage(client, "Login"))

	// 订阅个人频道
	go func() {
		c := models.Pool.Get()
		psc := redis.PubSubConn{Conn: c}
		_ = psc.Subscribe("appkey:" + unique)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				_, err := conn.Write([]byte(v.Data))
				if err != nil {
					log.Println("Send data error: ", err)
					return
				}
			case redis.Subscription:
			case error:
				log.Printf("Unknown type: %+v, %T", v, v)
				return
			}
		}
	}()

	// 向当前用户发送欢迎语
	c.Publish(unique, "Welcome to the GIM ChatRoom mode ^_^")

	// 接收当前用户输入数据
	go func() {
		buf := make([]byte, 2*1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				offline <- true
				if err == io.EOF {
					log.Println("Customer has closed the connection")
					return
				}
				log.Println("Recv data error: ", err)
				return
			}
			// 广播用户数据
			c.Publish("Broadcast", makeMessage(client, string(buf[:n])))
			online <- true
		}
	}()

	// 连接超时退出
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			c.delOnlineMap(unique)
			c.Publish("Broadcast", makeMessage(client, "Logout"))
			return
		// 超时退出
		case <-time.After(300 * time.Second):
			c.delOnlineMap(unique)
			c.Publish("Broadcast", makeMessage(client, "Time out"))
			return
		}
	}
}

// 基于消息推送的连接处理
func (m MessagePush) clusterHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := fmt.Sprintf("%s<->%s", client.Id, client.Addr)
	m.onlineMap[unique] = client

	// 向当前用户发送数据
	go func() {
		for msg := range client.C {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Send data error: ", err)
				return
			}
		}
	}()

	// 向当前用户发送欢迎语
	client.C <- makeMessage(client, "Welcome to the GIM MessagePush mode ^_^")

	// 连接超时退出
	for {
		select {
		case <-online:
		// 主动断开
		case <-offline:
			delete(m.onlineMap, unique)
			return
		}
	}
}
