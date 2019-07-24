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
	. "gim/src/models"
	"github.com/gomodule/redigo/redis"
	"log"
	"net"
	"strconv"
	"strings"
)

// 客户端
type Client struct {
	AppKey string      // 认证标识
	Id     string      // 客户端唯一ID, 由客户端维护该字段的唯一性
	Name   string      // 客户端名称
	City   string      // 城市
	Addr   string      // 客户端地址
	Mode   string      // 客户端模式
	C      chan string // 单播, 仅自己可见
}

// 基础数据结构
type Base struct {
	Mode      string            // 运行方式
	onlineMap map[string]Client // 在线队列
	Broadcast chan string       // 广播通道
}

// 聊天室模式
type ChatRoom struct {
	Base
}

// 消息推送模式
type MessagePush struct {
	Base
}

var (
	ChatRoomInstance    ChatRoom
	MessagePushInstance MessagePush
)

// 根据用户资料生成唯一ID
func (c Client) makeUniqueID() (unique string) {
	unique = strings.Join([]string{c.AppKey, c.Id}, ":")
	return
}

// 消息格式化: 公共广播
func makePublicMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("%s||[%s:%s] -> %s\n", client.AppKey, client.Id, client.Name, msg)
	return
}

// 消息格式化: 私有广播
func makePrivateMessage(client Client, msg string) (message string) {
	message = fmt.Sprintf("[%s:%s] -> %s\n", client.Id, client.Name, msg)
	return
}

// GIM 处理器
func GIMHandler(listener net.Listener, mode string) {
	// 监听公共广播频道, 解析数据处理
	switch mode {
	// 集群方式
	case "cluster":
		// 聊天室模式
		go func() {
			c := Pool.Get()
			psc := redis.PubSubConn{Conn: c}
			_ = psc.Subscribe("public:Broadcast")
			for {
				switch v := psc.Receive().(type) {
				case redis.Message:
					buf := strings.Split(string(v.Data), "||")
					// 获取在线列表
					if onlineMap, err := ChatRoomInstance.GetOnlineMap(buf[0]); err != nil {
						log.Printf("Get %s:onlineMap failed!\n", buf[0])
					} else {
						// 向在线用户广播数据
						for _, unique := range onlineMap {
							// 保存历史数据
							//ChatRoomInstance.SaveHistory([]string{unique}, buf[1])
							// 发往在线用户私人频道
							ChatRoomInstance.Publish(unique, buf[1], false)
						}
					}
				case redis.Subscription:
				case error:
					log.Printf("Unknown type: %+v, %T", v, v)
					return
				}
			}
		}()
	default:
		// 聊天室模式广播通道监听
		go func() {
			for {
				message := <-ChatRoomInstance.Broadcast
				// 遍历在线队列, 通知用户
				for _, client := range ChatRoomInstance.onlineMap {
					client.C <- message
				}
			}
		}()

		// 消息推送模式广播通道监听
		go func() {
			for {
				message := <-MessagePushInstance.Broadcast
				// 遍历在线队列, 通知用户
				for _, client := range MessagePushInstance.onlineMap {
					client.C <- message
				}
			}
		}()
	}

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
					client = Client{
						body[0], body[1], body[2], body[3],
						conn.RemoteAddr().String(), body[4], make(chan string),
					}
				}
				break
			}
		}
		switch client.Mode {
		// 聊天室模式连接处理
		case "chatroom":
			if ChatRoomInstance.Mode == "cluster" {
				go ChatRoomInstance.clusterHandler(conn, client)
			} else {
				go ChatRoomInstance.standaloneHandler(conn, client)
			}
			// 消息推送模式连接处理
		case "listener":
			if MessagePushInstance.Mode == "cluster" {
				go MessagePushInstance.clusterHandler(conn, client)
			} else {
				go MessagePushInstance.standaloneHandler(conn, client)
			}
		}
	}
}

func Run(mode string) {
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
		ChatRoomInstance = ChatRoom{Base{mode, make(map[string]Client), make(chan string)}}
		// 消息推送模式初始化
		MessagePushInstance = MessagePush{Base{mode, make(map[string]Client), make(chan string)}}
		// GIM 处理器实例化
		GIMHandler(listener, mode)
	}()
}
