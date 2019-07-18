/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         handler.go
@ Create Time:  2019/7/18 22:56
@ Software:     GoLand
*/

package im

import (
	"io"
	"log"
	"net"
	"time"
)

type User struct {
	Addr string      // 客户端地址
	C    chan string // 单播, 仅自己可见
}

// 聊天室
type ChatRoom struct {
	onlineMap map[string]User
	Broadcast chan string
}

// 接口封装
type Handler interface {
	Say()
}

// 消息格式化
//func makeMessage(client Client, msg string) (message string) {
//	message = fmt.Sprintf("[ %s ] -> %s", client.Addr, msg)
//	return
//}

// 基于聊天室的连接处理
func (c ChatRoom) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 客户端实例化
	clientAddress := conn.RemoteAddr().String()
	client := Client{clientAddress, make(chan string)}

	online, offline := make(chan bool), make(chan bool)

	// 加入在线队列
	onlineMap[clientAddress] = client
	// 广播用户上线
	Broadcast <- makeMessage(client, "Login")

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
			Broadcast <- makeMessage(client, string(buf[:n]))
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
			Broadcast <- makeMessage(client, "Logout")
			return
		// 超时退出
		case <-time.After(30000 * time.Second):
			delete(onlineMap, clientAddress)
			Broadcast <- makeMessage(client, "Time out")
			return
		}
	}
}
