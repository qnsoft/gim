/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-18 12:38
@ Software:     GoLand
*/

package main

import (
	_ "gim/src/im"
	. "gim/src/routers"
	"log"
)

func main() {
	// Start Restful API
	if err := App.Run(); err != nil {
		log.Fatalln("Service startup failed !", err)
	}
}
