/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         redis.go
@ Create Time:  2019-07-22 15:39
@ Software:     GoLand
*/

package models

import (
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type Redis struct {
	HOST        string `json:"host"`
	PORT        int    `json:"port"`
	DB          int    `json:"db"`
	MaxIdle     int    `json:"max_idle"`
	MaxActive   int    `json:"max_active"`
	IdleTimeout int    `json:"idle_timeout"`
}

var Pool *redis.Pool

func init() {
	Pool = &redis.Pool{
		MaxIdle:     Config.Redis.MaxIdle,
		MaxActive:   Config.Redis.MaxActive,
		IdleTimeout: time.Duration(Config.Redis.IdleTimeout) * time.Second,
		Wait:        true,
		Dial: func() (conn redis.Conn, err error) {
			conn, err = redis.Dial("tcp", strings.Join([]string{Config.Redis.HOST, strconv.Itoa(Config.Redis.PORT)}, ":"))
			if err != nil {
				return nil, err
			}
			_, _ = conn.Do("SELECT", Config.Redis.DB)
			return conn, nil
		},
	}

	// 首次启动清空仓库
	c := Pool.Get()
	_, _ = c.Do("FLUSHDB")
}
