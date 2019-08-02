/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         tokenizer.go
@ Create Time:  2019-08-02 17:42
@ Software:     GoLand
*/

package interfaces

import (
	"gim/app/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

type _params struct {
	AppKey    string `json:"app_key" form:"app_key" binding:"required"`
	AppSecret string `json:"app_secret" form:"app_secret" binding:"required"`
}

func Tokenizer(ctx *gin.Context) {
	var p _params
	if err := ctx.Bind(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "are you ok ?",
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"token": tools.GetMD5Hash(p.AppKey+p.AppSecret, false),
		})
	}
}
