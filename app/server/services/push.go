/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-29 18:42
@ Software:     GoLand
*/

package services

import (
	"encoding/json"
	. "gim/app/server/models"
	"log"
	"net"
	"strconv"
	"strings"
)

var MessagePushInstance MessagePush

func PushHandler(listener net.Listener) {
	// 监听公共消息队列
	go MessagePushInstance.Subscribe()

	// 监听请求连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Listener accept error: ", err)
			continue
		}

		// 初始化客户端
		var client Client
		buf := make([]byte, 1024)
		for {
			if n, _ := conn.Read(buf); n < 1024 {
				if err = json.Unmarshal(buf[:n], &client); err != nil {
					log.Println("Client initialization failed", err)
					return
				}
				// BUG: 验证 appkey 的有效性
				client.Addr = conn.RemoteAddr().String()
				client.C = make(chan string)
				break
			}
		}

		// 运行模式下发
		if _, err := conn.Write([]byte(MessagePushInstance.ServiceName)); err != nil {
			log.Println("Connection client exception", err)
			return
		}

		// 进入连接处理流程
		go MessagePushInstance.HandleConnection(conn, client)
	}
}

func init() {
	go func() {
		address := strings.Join([]string{Config.Services.Push.Host, strconv.Itoa(Config.Services.Push.Port)}, ":")
		// Start MessagePush service
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalln("MessagePush service startup failed", err)
		}
		log.Printf("MessagePush service starting TCP on: %s\n", address)

		defer listener.Close()

		// 初始化消息推送
		MessagePushInstance = MessagePush{Base{"push", make(map[string]Client)}}
		PushHandler(listener)
	}()
}
