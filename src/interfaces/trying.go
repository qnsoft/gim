/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         trying.go
@ Create Time:  2019-07-18 12:43
@ Software:     GoLand
*/

package interfaces

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Trying(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "Pong",
		"time": time.Now().Format("2006-01-02 15:04:05"),
	})
}
