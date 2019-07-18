/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-18 12:38
@ Software:     GoLand
*/

package main

import (
	"gim/src/im"
	. "gim/src/routers"
	"log"
)

func main() {
	// Start IM service
	go im.Run()
	// Start Restful API
	if err := App.Run(); err != nil {
		log.Println("Service startup failed !", err)
	}
}
