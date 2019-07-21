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
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// 客户端
type Client struct {
	Id   string      // 客户端唯一ID, 由客户端维护该字段的唯一性
	Name string      // 客户端名称
	City string      // 城市
	Addr string      // 客户端地址
	Mode string      // 客户端模式
	C    chan string // 单播, 仅自己可见
}

// 基础数据结构
type Base struct {
	onlineMap map[string]Client // 在线队列
	Broadcast chan string       // 广播通道
}

// 聊天室模式
type ChatRoom struct {
	Base
}

// 消息推送模式
type MessagePush struct {
	Base
}

// 消息格式化
func makeMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("[ %s:%s ] -> %s", client.Id, client.Name, msg)
	return
}

// GIM 处理器
func GIMHandler(listener net.Listener) {
	// 聊天室模式广播通道监听
	go func() {
		for {
			message := <-ChatRoomInstance.Broadcast
			// 遍历在线队列, 通知用户
			for _, client := range ChatRoomInstance.onlineMap {
				client.C <- message
			}
		}
	}()

	// 消息推送模式广播通道监听
	go func() {
		for {
			message := <-MessagePushInstance.Broadcast
			// 遍历在线队列, 通知用户
			for _, client := range MessagePushInstance.onlineMap {
				client.C <- message
			}
		}
	}()

	// 监听连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Listener accept error", err)
			continue
		}
		// 客户端实例化
		var client Client
		buf := make([]byte, 1024)
		for {
			n, _ := conn.Read(buf)
			if n < 1024 {
				if strings.HasPrefix(string(buf[:n]), "PROFILE:") {
					profile := strings.Split(string(buf[:n]), ":")[1]
					body := strings.Split(strings.ToLower(profile), "|")
					client = Client{
						body[0], body[1], body[2], conn.RemoteAddr().String(), body[3], make(chan string),
					}
				}
				break
			}
		}

		switch client.Mode {
		// 聊天室模式连接处理
		case "chatroom":
			go ChatRoomInstance.handleConnection(conn, client)
		// 消息推送模式连接处理
		case "listener":
			go MessagePushInstance.handleConnection(conn, client)
		}
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) handleConnection(conn net.Conn, client Client) {
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
func (m MessagePush) handleConnection(conn net.Conn, client Client) {
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
