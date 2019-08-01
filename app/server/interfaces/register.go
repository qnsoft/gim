/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         register.go
@ Create Time:  2019-08-01 11:00
@ Software:     GoLand
*/

package interfaces

import (
	"crypto/md5"
	"encoding/hex"
	"gim/app/server/models"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type RegisterParams struct {
	AppKey    string `json:"app_key" form:"app_key" binding:"required"`
	AppSecret string `json:"app_secret" form:"app_secret"`
	Title     string `json:"title" form:"title" binding:"required"`
	Reset     bool   `json:"reset" form:"reset"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GetMD5Hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

func Register(ctx *gin.Context) {
	var p RegisterParams
	if err := ctx.Bind(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"info": "Paramets error",
		})
	} else {
		p.AppKey = GetMD5Hash(p.AppKey)
		switch p.Reset {
		case true:
			p.AppSecret = GetMD5Hash(p.AppKey + StringWithCharset(4, charset))
		case false:
			p.AppSecret = GetMD5Hash(p.AppKey + "1990")
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
