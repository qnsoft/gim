/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-18 12:38
@ Software:     GoLand
*/

package main

import (
	"flag"
	"gim/src/im"
	. "gim/src/models"
	. "gim/src/routers"
	"log"
	"os"
)

var (
	help bool
	post string
	mode string
)

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.StringVar(&post, "port", "8080", "GIM restful api port")
	flag.StringVar(&mode, "mode", "cluster", "Start mode: [standalong, cluster]")
	flag.Parse()

	_ = os.Setenv("PORT", post)

	Config.Print()

	if help {
		flag.Usage()
	} else {
		// Start GIM server
		im.Run(mode)
		// Start Restful API
		if err := App.Run(); err != nil {
			log.Fatalln("Service startup failed !", err)
		}
	}
}
