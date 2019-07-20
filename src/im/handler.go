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
	"time"
)

// 客户端
type Client struct {
	Addr string      // 客户端地址, 由客户端维护该地址的唯一性
	C    chan string // 单播, 仅自己可见
}

// 聊天室
type ChatRoom struct {
	onlineMap map[string]Client // 在线队列
	Broadcast chan string       // 广播通道
}

// 消息格式化
func makeMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("[ %s ] -> %s", client.Addr, msg)
	return
}

// 基于聊天室的连接监听
func (c ChatRoom) listener(listener net.Listener) {
	// 监听广播通道
	go func() {
		for {
			message := <-c.Broadcast
			// 遍历在线队列, 通知用户
			for _, client := range c.onlineMap {
				client.C <- message
			}
		}
	}()

	// 监听连接请求
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Listener accept error", err)
			continue
		}
		// 处理连接
		go c.handleConnection(connection)
	}
}

// 基于聊天室的连接处理
func (c ChatRoom) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 客户端实例化
	clientAddress := conn.RemoteAddr().String()
	client := Client{clientAddress, make(chan string)}

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	c.onlineMap[clientAddress] = client
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
		// 异常退出
		case <-offline:
			delete(c.onlineMap, clientAddress)
			c.Broadcast <- makeMessage(client, "Logout")
			return
		// 超时退出
		case <-time.After(30000 * time.Second):
			delete(c.onlineMap, clientAddress)
			c.Broadcast <- makeMessage(client, "Time out")
			return
		}
	}
}
