/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         router.go
@ Create Time:  2019-07-18 12:37
@ Software:     GoLand
*/

package routers

import (
	. "gim/src/interfaces"
	"github.com/gin-gonic/gin"
	"net/http"
)

var App *gin.Engine

func init() {
	App = gin.Default()

	App.Any("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})

	App.GET("/trying", Trying)

	server := App.Group("/server")
	server.POST("/broadcast", Push)
}
