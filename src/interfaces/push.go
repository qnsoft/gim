/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-18 18:59
@ Software:     GoLand
*/

package interfaces

import (
	"gim/src/im"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Push(ctx *gin.Context) {
	im.Broadcast <- ctx.DefaultPostForm("message", "")
	ctx.String(http.StatusOK, "ok")
}
