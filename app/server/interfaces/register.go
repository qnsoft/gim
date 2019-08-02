/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         register.go
@ Create Time:  2019-08-01 11:00
@ Software:     GoLand
*/

package interfaces

import (
	"gim/app/server/models"
	"gim/app/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RegisterParams struct {
	AppKey    string `json:"app_key" form:"app_key" binding:"required"`
	AppSecret string `json:"app_secret"`
	Title     string `json:"title" form:"title" binding:"required"`
	Reset     bool   `json:"reset" form:"reset"`
}

func Register(ctx *gin.Context) {
	var p RegisterParams
	if err := ctx.Bind(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "Paramets error",
		})
	} else {
		p.AppKey = tools.GetMD5Hash(p.AppKey, false)
		switch p.Reset {
		case true:
			p.AppSecret = tools.GetMD5Hash(p.AppKey, true)
		case false:
			p.AppSecret = tools.GetMD5Hash(p.AppKey+"1990", false)
		}
		// 数据入库
		_, err := models.DB.Exec("insert into g_partners (app_key, app_secret, title) values (?, ?, ?)", p.AppKey, p.AppSecret, p.Title)
		if err != nil {
			switch p.Reset {
			case true:
				_, err := models.DB.Exec("update g_partners set app_secret=? where app_key=?", p.AppSecret, p.AppKey)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"info": "Data update failed",
					})
					return
				}
			case false:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"info": "Register failed",
				})
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"app_key":    p.AppKey,
			"app_secret": p.AppSecret,
			"title":      p.Title,
		})
	}
}
