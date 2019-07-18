/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         server.go
@ Create Time:  2019-07-18 12:49
@ Software:     GoLand
*/

package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// 客户端结构体
type Client struct {
	Addr string      // 客户端地址
	C    chan string // 单播, 只对自己可见
}

// 在线客户端队列
var onlineMap = make(map[string]Client)

// 广播通道
var broadcast = make(chan string)

// 消息格式化
func makeMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("[ %s ] -> %s", client.Addr, msg)
	return
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 客户端实例化
	clientAddress := conn.RemoteAddr().String()
	client := Client{clientAddress, make(chan string)}

	online := make(chan bool)
	offline := make(chan bool)

	// 加入在线队列
	onlineMap[clientAddress] = client
	// 广播用户上线
	broadcast <- makeMessage(client, "Login")

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
			broadcast <- makeMessage(client, string(buf[:n]))
			online <- true
		}
	}()

	// 连接超时退出
	for {
		select {
		case <-online:
		// 异常退出
		case <-offline:
			delete(onlineMap, clientAddress)
			broadcast <- makeMessage(client, "Logout")
			return
		// 超时退出
		case <-time.After(30 * time.Second):
			delete(onlineMap, clientAddress)
			broadcast <- makeMessage(client, "Time out")
			return
		}
	}
}

func Run() {
	host, port := "0.0.0.0", 8088
	// 启动IM服务端程序
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Println("IM startup failed !", err)
		return
	}
	log.Printf("IM service starting TCP on: %s:%d\n", host, port)

	defer listener.Close()

	// 监听广播通道
	go func() {
		for {
			message := <-broadcast
			// 遍历在线队列, 通知用户
			for _, client := range onlineMap {
				client.C <- message
			}
		}
	}()

	// 监听客户端连接
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Listener accept error", err)
			continue
		}
		// 处理连接
		go handleConnection(connection)
	}
}
