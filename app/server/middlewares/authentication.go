/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         authentication.go
@ Create Time:  2019-08-01 17:21
@ Software:     GoLand
*/

package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Auth struct {
	Token string `json:"token" form:"token" binding:"required"`
}

func Authentication(ctx *gin.Context) {
	var p Auth
	if err := ctx.ShouldBind(&p); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"info": "Invalid token",
		})
		ctx.Abort()
	}
}
