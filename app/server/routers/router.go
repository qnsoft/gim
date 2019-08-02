/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         router.go
@ Create Time:  2019-07-29 14:48
@ Software:     GoLand
*/

package routers

import (
	"fmt"
	"gim/app/server/interfaces"
	. "gim/app/server/middlewares"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"time"
)

var App *gin.Engine

func init() {
	switch os.Getenv("GIN_MODE") {
	case "release":
		f, _ := os.Create("app/logs/access.log")
		gin.DisableConsoleColor()
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	default:
	}

	App = gin.New()
	App.Use(gin.Recovery())
	App.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %d %s \"%s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.BodySize,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	App.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"info": "Hello world",
			"time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	App.POST("/token", interfaces.Tokenizer)
	App.POST("/register", interfaces.Register)

	services := App.Group("/service", Authentication)
	services.POST("/push", interfaces.Push)
}
