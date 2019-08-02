/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         authentication.go
@ Create Time:  2019-08-01 17:21
@ Software:     GoLand
*/

package middlewares

import (
	"gim/app/server/models"
	"gim/app/tools"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

type Auth struct {
	AppKey string `json:"appkey" form:"appkey" binding:"required"`
	Token  string `json:"token" form:"token" binding:"required"`
}

func Validate(appKey, token string) bool {
	c := models.Pool.Get()
	defer c.Close()
	// 读取缓存
	appSecret, err := redis.String(c.Do("GET", "appkey:"+appKey))
	if err != nil {
		log.Println("Redis->GET: Validate failed", err)
	}
	// 读取数据库
	if appSecret == "" {
		sql := "select app_secret from g_partners where app_key=?"
		row := models.DB.QueryRow(sql, appKey)
		if err = row.Scan(&appSecret); err != nil {
			log.Println("Mysql->QueryRow Scan: Validate failed", err)
			return false
		}
		// 加入缓存
		_, _ = c.Do("SET", "appkey:"+appKey, appSecret)
	}
	// 校验
	if token == tools.GetMD5Hash(appKey+appSecret, false) {
		return true
	}
	return false
}

// 接口认证
func Authentication(ctx *gin.Context) {
	var p Auth
	if err := ctx.ShouldBind(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "Paramets error",
		})
		ctx.Abort()
	} else if !Validate(p.AppKey, p.Token) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"info": "Invalid token",
		})
		ctx.Abort()
	}
}
