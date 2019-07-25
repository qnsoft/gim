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
	appkey   string
	id       string
	name     string
	city     string
	retry    int
	interval int
	loop     bool
)

// 客户端数据结构
type Client struct {
	AppKey string
	Id     string
	Name   string
	City   string
}

// 公共处理函数
func (c Client) Handler(retry, interval int, loop bool) {
	for try := 0; try < retry; try++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 3*time.Second)
		if err != nil {
			<-time.After(time.Duration(try*interval) * time.Second)
			log.Println("Trying to reconnect...")

			if loop {
				try--
				<-time.After(time.Duration(interval) * time.Second)
			}
			continue
		}

		// 发送基础数据
		if _, err := conn.Write([]byte(fmt.Sprintf("PROFILE:%s|%s|%s|%s", c.AppKey, c.Id, c.Name, c.City))); err != nil {
			log.Println("Send profile failed: ", err)
			return
		}

		// 获取运行模式
		buf := make([]byte, 1024)
		if n, err := conn.Read(buf); err != nil {
			log.Println("Unable to get mode!", err)
			return
		} else {
			switch string(buf[:n]) {
			case "im":
				Im(conn)
			case "push":
				Push(conn)
			default:
				log.Fatalf("No matching pattern.")
			}
		}
	}
	log.Println("Unable to connect to im server.")
}

// 聊天室模式
func Im(conn net.Conn) {
	defer conn.Close()

	online, closed := make(chan bool), make(chan bool)
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
			if _, err = conn.Write(buf[:n]); err != nil {
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

// 消息推送模式
func Push(conn net.Conn) {
	defer conn.Close()

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
	flag.IntVar(&port, "port", 8081, "GIM server listener port")
	flag.StringVar(&appkey, "appkey", "", "AppKey")
	flag.StringVar(&id, "id", "0827", "Client unique id")
	flag.StringVar(&name, "name", "guest", "Client name")
	flag.StringVar(&city, "city", "BJ", "Client city name")
	flag.IntVar(&retry, "retry", 3, "Number of connection retries")
	flag.IntVar(&interval, "interval", 3, "Connection retry interval")
	flag.BoolVar(&loop, "loop", false, "Infinite retry")

	flag.Parse()

	if help {
		flag.Usage()
	} else {
		// 客户端实例化
		client := Client{appkey, id, name, city}
		// 断线重连次数、间隔、回调函数 模式
		client.Handler(retry, interval, loop)
	}
}
