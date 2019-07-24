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
	AppKey string `form:"appkey" json:"app_key" binding:"required"`
	Mode   string `form:"mode" json:"mode" binding:"required"`
	Msg    string `form:"msg" json:"msg" binding:"required"`
}

func Push(ctx *gin.Context) {
	var params Params
	if err := ctx.Bind(&params); err != nil {
		ctx.Status(http.StatusBadRequest)
	}
	switch params.Mode {
	case "chatroom":
		switch ChatRoomInstance.Mode {
		case "cluster":
			if onlineMap, err := ChatRoomInstance.GetOnlineMap(params.AppKey); err != nil {
				ctx.Status(http.StatusBadRequest)
			} else {
				for _, unique := range onlineMap {
					ChatRoomInstance.Publish(unique, params.Msg, false)
				}
				ctx.Status(http.StatusOK)
			}
		default:
			ChatRoomInstance.Broadcast <- fmt.Sprintf("[ Game ] -> %s", params.Msg)
		}
	case "listener":
		switch MessagePushInstance.Mode {
		case "cluster":
			if onlineMap, err := MessagePushInstance.GetOnlineMap(params.AppKey); err != nil {
				ctx.Status(http.StatusBadRequest)
			} else {
				for _, unique := range onlineMap {
					MessagePushInstance.Publish(unique, params.Msg, false)
				}
				ctx.Status(http.StatusOK)
			}
		default:
			MessagePushInstance.Broadcast <- fmt.Sprintf("[ Game ] -> %s", params.Msg)
		}
	}
	ctx.Status(http.StatusOK)
}
