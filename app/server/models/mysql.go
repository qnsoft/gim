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
	"log"
)

type Mysql struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       string `json:"db"`
	User     string `json:"user"`
	Password string `json:"password"`
}

var DB *sql.DB

func Insert() bool {
	tx, err := DB.Begin()
	if err != nil {
		log.Println("starting transaction failed", err)
		return false
	}
	stmt, err := tx.Prepare("INSERT INTO abc (`name`, `token`) VALUES (?, ?)")
	if err != nil {
		log.Println("Prepare sql failed", err)
		return false
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Println("Exec failed", err)
		return false
	}
	_ = tx.Commit()
	return true
}

func init() {
	// e.g. user:password@tcp(host:port)/database?charset=utf8
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", Config.Mysql.User, Config.Mysql.Password, Config.Mysql.Host, Config.Mysql.Port, Config.Mysql.DB)
	DB, _ = sql.Open("mysql", uri)
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		log.Println("Open database failed", err)
		return
	}
	log.Println("Database connection succeeded")
}
