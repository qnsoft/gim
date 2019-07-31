/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         register.go
@ Create Time:  2019-07-31 18:13
@ Software:     GoLand
*/

package interfaces

import (
	"fmt"
	"gim/app/server/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type params struct {
	Token  string `json:"token" form:"token" binding:"required"`
	AppKey string `json:"appkey" form:"appkey" binding:"required"`
	Name   string `json:"name" form:"name" binding:"required"`
}

func Register(ctx *gin.Context) {
	var params params
	if err := ctx.Bind(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "Parameter error",
		})
	} else {
		fmt.Println(params)
		models.Insert()
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "Hello world",
		})
	}
}
