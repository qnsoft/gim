/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-29 16:02
@ Software:     GoLand
*/

package main

import (
	"encoding/json"
	"flag"
	"gim/app/tools"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	help      bool
	host      string
	port      int
	appKey    string
	appSecret string
	id        string
	name      string
	city      string
	retry     int
	interval  int
	loop      bool
)

type Client struct {
	AppKey    string `json:"app_key"`    // 认证标识
	AppSecret string `json:"app_secret"` // 安全码
	Token     string `json:"token"`      // 认证令牌
	Id        string `json:"id"`         // 客户端唯一ID, 由客户端维护该字段的唯一性
	Name      string `json:"name"`       // 客户端名称
	City      string `json:"city"`       // 城市
}

func Tokenizer() string {
	return tools.GetMD5Hash(appKey+appSecret, false)
}

func im(conn net.Conn) {
	defer conn.Close()

	online, closed := make(chan bool), make(chan bool)

	// 接收输出
	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := conn.Read(buf); err != nil {
				closed <- true
				log.Println("Connection is closed")
				return
			} else {
				online <- true
				log.Println(string(buf[:n]))
			}
		}
	}()

	// 接收输入
	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := os.Stdin.Read(buf); err != nil {
				log.Println("Receive terminal input error", err)
				continue
			} else {
				if _, err := conn.Write(buf[:n]); err != nil {
					closed <- true
					log.Println("Failed to send data", err)
					return
				}
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

func push(conn net.Conn) {
	defer conn.Close()

	// 接收输出
	buf := make([]byte, 1024)
	for {
		if n, err := conn.Read(buf); err != nil {
			log.Println("Connection is closed")
			return
		} else {
			log.Println(string(buf[:n]))
		}
	}
}

func (c Client) Handler(retry, interval int, loop bool) {
	for try := 0; try < retry; try++ {
		if loop {
			try, interval = 1, 5
		}
		conn, err := net.DialTimeout("tcp", strings.Join([]string{host, strconv.Itoa(port)}, ":"), 3*time.Second)
		if err != nil {
			log.Println("Trying to reconnect...", err)
			<-time.After(time.Duration(try*interval) * time.Second)
			continue
		}

		// 客户端认证
		var mode string
		if buf, err := json.Marshal(c); err != nil {
			log.Panic(err)
		} else {
			if _, err := conn.Write(buf); err != nil {
				log.Panic(err)
			} else {
				buf := make([]byte, 1024)
				if n, err := conn.Read(buf); err != nil {
					log.Println("Connection is closed")
					return
				} else {
					mode = string(buf[:n])
				}
			}
		}

		// Heartbeat detection
		go func() {
			for {
				if _, err := conn.Write([]byte("Heartbeat:ack")); err != nil {
					return
				}
				<-time.After(300 * time.Second)
			}
		}()

		switch mode {
		case "im":
			im(conn)
		case "push":
			push(conn)
		default:
			log.Fatalln("No matching pattern")
		}
	}
	log.Println("Unable to connect to GIM server")
}

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.StringVar(&host, "host", "127.0.0.1", "GIM server host")
	flag.IntVar(&port, "port", 8081, "GIM server port")
	flag.StringVar(&appKey, "appkey", "", "Authorized appkey")
	flag.StringVar(&appSecret, "appsecret", "", "Authorized appsecret")
	flag.StringVar(&id, "id", "001", "Client unique id")
	flag.StringVar(&name, "name", "guest", "Client name")
	flag.StringVar(&city, "city", "bj", "Client city name")
	flag.IntVar(&retry, "retry", 3, "Number of failed retries")
	flag.IntVar(&interval, "interval", 5, "Failure retry interval -> second")
	flag.BoolVar(&loop, "loop", false, "Infinite retry")
	flag.Parse()

	if help {
		flag.Usage()
	} else {
		// 初始化客户端
		client := Client{AppKey: appKey, Token: Tokenizer(), Id: id, Name: name, City: city}
		client.Handler(retry, interval, loop)
	}
}
