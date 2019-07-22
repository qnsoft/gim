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
	"time"
)

type Redis struct {
	HOST string `json:"host"`
	PORT int    `json:"port"`
	//DB          int    `json:"db"`
	//MaxIdle     int    `json:"max_idle"`
	//MaxActive   int    `json:"max_active"`
	//IdleTimeout int    `json:"idle_timeout"`
}

var Pool *redis.Pool

func init() {
	Pool = &redis.Pool{
		MaxIdle:     1,
		MaxActive:   0,
		IdleTimeout: 30 * time.Second,
		Dial: func() (conn redis.Conn, err error) {
			conn, err = redis.Dial("tcp", "0.0.0.0:6379")
			if err != nil {
				return nil, err
			}
			_, _ = conn.Do("select", 1)
			return conn, nil
		},
	}
}
