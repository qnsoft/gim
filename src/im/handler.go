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
	"strings"
	"time"
)

// 集群模式在线列表: 增加
func (b Base) addOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	_, err := c.Do("SADD", appKey+":onlineMap", id)
	if err != nil {
		log.Println("SADD failed!", err)
	}
}

// 集群模式在线列表: 移除
func (b Base) delOnlineMap(id, appKey string) {
	c := models.Pool.Get()
	defer c.Close()
	_, err := c.Do("SREM", appKey+":onlineMap", id)
	if err != nil {
		log.Println("SREM failed!", err)
	}
}

// 集群模式在线列表: 获取集合数据
func (b Base) GetOnlineMap(appKey string) ([]string, error) {
	c := models.Pool.Get()
	defer c.Close()
	_ = c.Send("SMEMBERS", appKey+":onlineMap")
	_ = c.Flush()
	if reply, err := redis.Strings(c.Receive()); err != nil {
		log.Println("SMEMBERS failed!", err)
		return nil, err
	} else {
		return reply, nil
	}
}

// 集群模式离线消息: 存储
func (b Base) SaveHistory(ids []string, msg interface{}) {
	c := models.Pool.Get()
	defer c.Close()
	for _, unique := range ids {
		_ = c.Send("RPUSH", unique, msg)
	}
	if err := c.Flush(); err != nil {
		log.Println("SaveHistory faield!", err)
	}
}

// 集群模式: 频道发布
func (b Base) Publish(channel, msg string, public bool) {
	c := models.Pool.Get()
	defer c.Close()
	switch public {
	// 发给公共频道 -> public:Broadcast
	case true:
		if _, err := c.Do("PUBLISH", "public:"+channel, msg); err != nil {
			log.Println("PUBLISH failed", err)
		}
	// 发给私有频道 -> appkey:unique
	default:
		if _, err := c.Do("PUBLISH", channel, msg); err != nil {
			log.Println("PUBLISH failed", err)
		}
	}
}

// 集群模式: 频道订阅
func (b Base) Subscribe() {

}

// 基于聊天室的连接处理: Standalone
func (c ChatRoom) standaloneHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := fmt.Sprintf("%s<->%s", client.Id, client.Addr)
	c.onlineMap[unique] = client
	// 广播用户上线
	c.Broadcast <- makePrivateMessage(client, "Login")

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
	client.C <- makePrivateMessage(client, "Welcome to the GIM ChatRoom mode ^_^")

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
			c.Broadcast <- makePrivateMessage(client, string(buf[:n]))
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
			c.Broadcast <- makePrivateMessage(client, "Logout")
			return
		// 超时退出
		case <-time.After(300 * time.Second):
			delete(c.onlineMap, unique)
			c.Broadcast <- makePrivateMessage(client, "Time out")
			return
		}
	}
}

// 基于消息推送的连接处理: Standalone
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
	client.C <- makePrivateMessage(client, "Welcome to the GIM MessagePush mode ^_^")

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

// 基于聊天室的连接处理: Cluster
func (c ChatRoom) clusterHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := strings.Join([]string{client.AppKey, client.Id}, ":")
	c.addOnlineMap(unique, client.AppKey)

	// 广播用户上线
	c.Publish("Broadcast", makePublicMessage(client, "Login"), true)

	// 订阅私人频道, 接收发给自己的数据
	go func() {
		c := models.Pool.Get()
		psc := redis.PubSubConn{Conn: c}
		_ = psc.Subscribe(unique)
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

	// 向私人频道发送欢迎语
	c.Publish(unique, "Welcome to the GIM ChatRoom mode ^_^", false)

	// 集群模式离线消息: 读取
	go func() {
		c := models.Pool.Get()
		defer c.Close()
		_ = c.Send("LRANGE", unique, 0, -1)
		_ = c.Flush()
		for {
			reply, err := redis.ByteSlices(c.Receive())
			if err != nil {
				log.Println("SMEMBERS failed!", err)
				return
			}
			for _, msg := range reply {
				if _, err = conn.Write(msg); err != nil {
					log.Println("Send data error: ", err)
					return
				}
			}
		}
	}()

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
			c.Publish("Broadcast", makePublicMessage(client, string(buf[:n])), true)
			online <- true
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

// 基于消息推送的连接处理: Cluster
func (m MessagePush) clusterHandler(conn net.Conn, client Client) {
	defer conn.Close()

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	unique := strings.Join([]string{client.AppKey, client.Id}, ":")
	m.addOnlineMap(unique, client.AppKey)

	// 订阅私人频道, 接收发给自己的数据
	go func() {
		c := models.Pool.Get()
		psc := redis.PubSubConn{Conn: c}
		_ = psc.Subscribe(unique)
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
	m.Publish(unique, "Welcome to the GIM MessagePush mode ^_^", false)

	// 连接超时退出
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
