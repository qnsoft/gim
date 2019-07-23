/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         server.go
@ Create Time:  2019-07-18 12:49
@ Software:     GoLand
*/

package im

import (
	. "gim/src/models"
	"log"
	"net"
	"strconv"
	"strings"
)

var (
	ChatRoomInstance    ChatRoom
	MessagePushInstance MessagePush
)

func Run() {
	go func() {
		address := strings.Join([]string{Config.Server.HOST, strconv.Itoa(Config.Server.PORT)}, ":")
		// 启动IM服务端程序
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalln("IM service startup failed !", err)
		}
		log.Printf("IM service starting TCP on: %s\n", address)

		defer listener.Close()

		// 聊天室模式初始化
		ChatRoomInstance = ChatRoom{Base{make(map[string]Client), make(chan string)}}
		// 消息推送模式初始化
		MessagePushInstance = MessagePush{Base{make(map[string]Client), make(chan string)}}
		// GIM 处理器实例化
		GIMHandler(listener)
	}()
}
