/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         mysql.go
@ Create Time:  2019-07-31 16:29
@ Software:     GoLand
*/

package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes/timestamp"
	"log"
)

type Mysql struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       string `json:"db"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Table: g_partners
type Partners struct {
	Id        int                 `json:"id"`
	AppKey    string              `json:"app_key"`
	AppSecret string              `json:"app_secret"`
	Title     string              `json:"title"`
	CreatedAt timestamp.Timestamp `json:"created_at"`
	UpdatedAt timestamp.Timestamp `json:"updated_at"`
}

// Table: g_clients
type Clients struct {
	Id        int                 `json:"id"`
	CId       string              `json:"c_id"`
	CName     string              `json:"c_name"`
	CCity     string              `json:"c_city"`
	CAddr     string              `json:"c_addr"`
	CreatedAt timestamp.Timestamp `json:"created_at"`
	UpdatedAt timestamp.Timestamp `json:"updated_at"`
}

var DB *sql.DB

func init() {
	// e.g. user:password@tcp(host:port)/database?charset=utf8
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", Config.Mysql.User, Config.Mysql.Password, Config.Mysql.Host, Config.Mysql.Port, Config.Mysql.DB)
	DB, _ = sql.Open("mysql", uri)
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		log.Println("Mysql: Open database failed, ", err)
		return
	}
	log.Println("Mysql: Connection established successfully !")
}
