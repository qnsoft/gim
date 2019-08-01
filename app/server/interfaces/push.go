/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-31 17:42
@ Software:     GoLand
*/

package interfaces

import (
	"encoding/json"
	. "gim/app/server/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PushParams struct {
	AppKey  string `form:"appkey" json:"appkey" binding:"required"`
	Mode    string `form:"mode" json:"mode" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
}

func Push(ctx *gin.Context) {
	var p PushParams
	if err := ctx.Bind(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "Paramets error",
		})
	} else {
		// TODO 定向推送
		buf, _ := json.Marshal(PublicMessage{AppKey: p.AppKey, To: "all", Content: p.Message})
		switch p.Mode {
		case "im":
			ChatRoomInstance.Publish(string(buf))
		case "push":
			MessagePushInstance.Publish(string(buf))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"info": "ok",
		})
	}
}
