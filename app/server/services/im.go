/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         im.go
@ Create Time:  2019-07-29 14:54
@ Software:     GoLand
*/

package services

import (
	"gim/app/server/middlewares"
	. "gim/app/server/models"
	"log"
	"net"
	"strconv"
	"strings"
)

var ChatRoomInstance ChatRoom

func ChatRoomHandler(listener net.Listener) {
	// 监听公共消息队列
	go ChatRoomInstance.Subscribe()

	// 监听请求连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("IM: Listening for connection failure, ", err)
			continue
		}

		// 初始化客户端
		client, err := InitClient(conn)
		if err != nil {
			log.Println("IM: initClient failed, ", err)
			conn.Close()
			continue
		}

		// 客户端有效性检验
		if !middlewares.Validate(client.AppKey, client.Token) {
			conn.Close()
			continue
		}

		// 运行模式下发
		if _, err := conn.Write([]byte(ChatRoomInstance.ServiceName)); err != nil {
			log.Println("IM: Send ChatRoomInstance.ServiceName failed, ", err)
			conn.Close()
			continue
		}

		// 进入连接处理流程
		go ChatRoomInstance.HandleConnection(conn, client)
	}
}

func init() {
	go func() {
		address := strings.Join([]string{Config.Services.IM.Host, strconv.Itoa(Config.Services.IM.Port)}, ":")
		// Start IM service
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalln("IM: Service startup failed, ", err)
		}
		log.Println("IM: Service starting tcp on: ", address)

		defer listener.Close()

		// 初始化聊天室
		ChatRoomInstance = ChatRoom{Base{"im", make(map[string]Client)}}
		ChatRoomHandler(listener)
	}()
}
