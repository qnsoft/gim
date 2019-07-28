/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019/7/24 22:09
@ Software:     GoLand
*/

package services

import (
	. "gim/src/models"
	"github.com/gomodule/redigo/redis"
	"log"
	"net"
	"strconv"
	"strings"
)

var MessagePushInstance MessagePush

// GIM 处理器
func PushHandler(listener net.Listener) {
	// 监听公共广播频道, 解析数据处理
	go func() {
		c := Pool.Get()
		psc := redis.PubSubConn{Conn: c}
		_ = psc.Subscribe(strings.Join([]string{MessagePushInstance.ServiceName, "public", "Broadcast"}, ":"))
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				buf := strings.Split(string(v.Data), "||")
				// 获取在线列表
				if onlineMap, err := MessagePushInstance.GetOnlineMap(buf[0]); err != nil {
					log.Printf("Get %s:onlineMap failed!\n", buf[0])
				} else {
					// 向在线用户广播数据
					for _, unique := range onlineMap {
						// 发往在线用户私人频道
						privateChannel := strings.Join([]string{MessagePushInstance.ServiceName, unique}, ":")
						MessagePushInstance.Publish(privateChannel, buf[1], false)
					}
				}
			case redis.Subscription:
			case error:
				log.Printf("Unknown type: %+v, %T", v, v)
				return
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
					// BUG: 必须验证 appkey
					client = Client{
						body[0], body[1], body[2], body[3], conn.RemoteAddr().String(),
					}
				}
				break
			}
		}

		// 下发监听模式
		if _, err = conn.Write([]byte(MessagePushInstance.ServiceName)); err != nil {
			log.Println("Send runing mode failed.", err)
			return
		}

		// 聊天室模式连接处理
		go MessagePushInstance.HandleConnection(conn, client)
	}
}

func init() {
	go func() {
		address := strings.Join([]string{Config.Services.Push.HOST, strconv.Itoa(Config.Services.Push.PORT)}, ":")
		// 启动IM服务端程序
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalln("Push service startup failed !", err)
		}
		log.Printf("Push service starting TCP on: %s\n", address)

		defer listener.Close()

		// 消息推送模式初始化
		MessagePushInstance = MessagePush{Base{"push"}}
		// GIM 处理器实例化
		PushHandler(listener)
	}()
}
