/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         server.go
@ Create Time:  2019-07-18 12:49
@ Software:     GoLand
*/

package im

import (
	"fmt"
	"log"
	"net"
)

var ChatRoomInstance ChatRoom

func init() {
	go func() {
		host, port := "0.0.0.0", 8088
		// 启动IM服务端程序
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			log.Fatalln("IM service startup failed !", err)
		}
		log.Printf("IM service starting TCP on: %s:%d\n", host, port)

		defer listener.Close()

		// 聊天室实例化
		ChatRoomInstance = ChatRoom{make(map[string]Client), make(chan string)}
		ChatRoomInstance.listener(listener)
	}()
}
