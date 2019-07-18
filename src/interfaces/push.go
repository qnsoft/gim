/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-18 18:59
@ Software:     GoLand
*/

package interfaces

import (
	"gim/src/server"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Push(ctx *gin.Context) {
	server.Broadcast <- ctx.PostForm("message")
	ctx.String(http.StatusOK, "ok")
}
