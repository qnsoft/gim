/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-29 14:44
@ Software:     GoLand
*/

package main

import (
	"flag"
	. "gim/app/server/routers"
	_ "gim/app/server/services"
	"log"
	"os"
	"strconv"
)

var (
	help bool
	port int
)

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.IntVar(&port, "port", 8080, "GIM restful api port")
	flag.Parse()

	_ = os.Setenv("PORT", strconv.Itoa(port))

	if help {
		flag.Usage()
	} else {
		// Start Restful API
		if err := App.Run(); err != nil {
			log.Fatalln("Service startup failed.", err)
		}
	}
}
