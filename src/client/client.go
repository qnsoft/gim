/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         client.go
@ Create Time:  2019-07-18 13:56
@ Software:     GoLand
*/

package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// 与服务端建立连接
	connection, err := net.Dial("tcp", "0.0.0.0:8088")
	if err != nil {
		log.Panicln("Unable to connect to server: ", err)
		return
	}
	log.Println("Connection succeeded.")

	defer connection.Close()

	// 接收服务器返回数据
	go func() {
		buf := make([]byte, 2*1024)
		for {
			n, err := connection.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("Connection is closed.")
					return
				}
				log.Println("Recv data error: ", err)
				return
			}
			log.Println(string(buf[:n]))
		}
	}()

	// 监听用户输入, 向服务器发送数据
	buf := make([]byte, 1024)
	for {
		// 监听标准输入
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Println("Stdin error: ", err)
			return
		}
		_, err = connection.Write(buf[:n])
		if err != nil {
			log.Println("Send data error: ", err)
			return
		}
	}
}
