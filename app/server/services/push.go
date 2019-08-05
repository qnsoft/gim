/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-29 18:42
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

var MessagePushInstance MessagePush

func PushHandler(listener net.Listener) {
	// 监听公共消息队列
	go MessagePushInstance.Subscribe()

	// 监听请求连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("PUSH: Listening for connection failure, ", err)
			continue
		}

		// 初始化客户端
		client, err := InitClient(conn)
		if err != nil {
			log.Println("PUSH: initClient failed, ", err)
			conn.Close()
			continue
		}

		// 客户端有效性检验
		if !middlewares.Validate(client.AppKey, client.Token) {
			conn.Close()
			continue
		}

		// 运行模式下发
		if _, err := conn.Write([]byte(MessagePushInstance.ServiceName)); err != nil {
			log.Println("PUSH: Send MessagePushInstance.ServiceName failed, ", err)
			conn.Close()
			continue
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
			log.Fatalln("PUSH: Service startup failed, ", err)
		}
		log.Println("PUSH: Service starting tcp on: ", address)

		defer listener.Close()

		// 初始化消息推送
		MessagePushInstance = MessagePush{Base{"push", make(map[string]Client)}}
		PushHandler(listener)
	}()
}
