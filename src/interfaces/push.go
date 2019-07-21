/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-18 18:59
@ Software:     GoLand
*/

package interfaces

import (
	"fmt"
	. "gim/src/im"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Params struct {
	Mode string `form:"mode" json:"mode" binding:"required"`
	Msg  string `form:"msg" json:"msg" binding:"required"`
}

func Push(ctx *gin.Context) {
	var params Params
	if err := ctx.Bind(&params); err != nil {
		ctx.Status(http.StatusBadRequest)
	}
	switch params.Mode {
	case "chatroom":
		ChatRoomInstance.Broadcast <- fmt.Sprintf("[ Game ] -> %s", params.Msg)
	case "listener":
		MessagePushInstance.Broadcast <- fmt.Sprintf("[ Game ] -> %s", params.Msg)
	}
	ctx.Status(http.StatusOK)
}
