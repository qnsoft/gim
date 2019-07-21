/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         client.go
@ Create Time:  2019-07-18 13:56
@ Software:     GoLand
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	help     bool
	host     string
	port     int
	id       string
	name     string
	city     string
	retry    int
	interval int
	mode     string
	callback Callback
)

type Client struct {
	Id   string
	Name string
	City string
	Mode string
}

type Callback func(conn net.Conn, client Client)

func (c Client) Handler(retry, interval int, callback Callback) {
	for try := 0; try < retry; try++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 3*time.Second)
		if err != nil {
			<-time.After(time.Duration(try*interval) * time.Second)
			log.Printf("Trying to reconnect %d...", try+1)
			continue
		}
		callback(conn, c)
	}
	log.Println("Unable to connect to im server.")
}

// 聊天室模式
func ChatRoom(conn net.Conn, client Client) {
	defer conn.Close()

	online, closed := make(chan bool), make(chan bool)

	// 发送基础数据
	_, err := conn.Write([]byte(fmt.Sprintf("PROFILE:%s|%s|%s|%s", client.Id, client.Name, client.City, client.Mode)))
	if err != nil {
		log.Println("Send profile failed: ", err)
	}

	// 接收服务器返回数据
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				closed <- true
				log.Println("Connection is closed.")
				return
			}
			online <- true
			log.Println(string(buf[:n]))
		}
	}()

	// 监听用户输入, 向服务器发送数据
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Println("Stdin error: ", err)
				continue
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				closed <- true
				log.Println("Send data error: ", err)
				return
			}
		}
	}()

	for {
		select {
		case <-online:
		case <-closed:
			return
		}
	}
}

// 被动监听模式
func Listener(conn net.Conn, client Client) {
	defer conn.Close()

	// 发送基础数据
	_, err := conn.Write([]byte(fmt.Sprintf("PROFILE:%s|%s|%s|%s", client.Id, client.Name, client.City, client.Mode)))
	if err != nil {
		log.Println("Send profile failed: ", err)
		return
	}

	// 接收服务器返回数据
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Connection is closed.")
			return
		}
		log.Println(string(buf[:n]))
	}
}

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.StringVar(&host, "host", "127.0.0.1", "GIM server address")
	flag.IntVar(&port, "port", 8088, "GIM server listener port")
	flag.StringVar(&id, "id", "0827", "Client unique id")
	flag.StringVar(&name, "name", "guest", "Client name")
	flag.StringVar(&city, "city", "BJ", "Client city name")
	flag.IntVar(&retry, "retry", 3, "Number of connection retries")
	flag.IntVar(&interval, "interval", 1, "Connection retry interval")
	flag.StringVar(&mode, "mode", "chatroom", "Access mode, [chatroom, listener]")

	flag.Parse()

	if help {
		flag.Usage()
	} else {
		// 客户端实例化
		client := Client{id, name, city, mode}
		// 模式判断
		switch mode {
		case "chatroom":
			callback = ChatRoom
		case "listener":
			callback = Listener
		default:
			callback = ChatRoom
		}
		// 断线重连次数、间隔、回调函数 模式
		client.Handler(retry, interval, callback)
	}
}
