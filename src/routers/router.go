/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         router.go
@ Create Time:  2019-07-18 12:37
@ Software:     GoLand
*/

package routers

import (
	"gim/src/interfaces"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var App *gin.Engine

func init() {
	App = gin.Default()

	App.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message":  "Hello world",
			"datetime": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	Gim := App.Group("/services")
	Gim.POST("/push", interfaces.Push)
}
