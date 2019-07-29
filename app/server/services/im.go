/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         im.go
@ Create Time:  2019-07-29 14:54
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

var ChatRoomInstance ChatRoom

func ChatRoomHandler(listener net.Listener) {
	// 监听公共消息队列
	go func() {

	}()

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
		if _, err := conn.Write([]byte(ChatRoomInstance.ServiceName)); err != nil {
			log.Println("Connection client exception", err)
			return
		}

		// 进入连接处理流程
		go ChatRoomInstance.HandleConnection(conn, client)
	}
}

func init() {
	go func() {
		address := strings.Join([]string{Config.Services.IM.HOST, strconv.Itoa(Config.Services.IM.PORT)}, ":")
		// Start IM service
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalln("IM service startup failed", err)
		}
		log.Printf("IM service starting TCP on: %s\n", address)

		defer listener.Close()

		// 初始化聊天室
		ChatRoomInstance = ChatRoom{Base{"im", make(map[string]Client)}}
		ChatRoomHandler(listener)
	}()
}
