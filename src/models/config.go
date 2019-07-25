/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         config.go
@ Create Time:  2019-07-22 18:46
@ Software:     GoLand
*/

package models

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var Config Conf

// Default
type Public struct {
	HOST string `json:"host"`
	PORT int    `json:"port"`
}

// IM Server
type Im struct {
	Public `json:"im"`
}

// Push Server
type Push struct {
	Public `json:"push"`
}

// Services
type Services struct {
	Im
	Push
}

// 配置文件结构体
type Conf struct {
	Redis    Redis
	Services Services
}

// 自定义json格式化打印方法
func (c Conf) Print() {
	buf, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		log.Println("Error: ", err)
	}
	log.Println(string(buf))
}

// 加载配置文件
func init() {
	buf, err := ioutil.ReadFile("src/config.json")
	if err != nil {
		log.Fatalf("Unable to load configuration file: %v", err)
	}
	err = json.Unmarshal(buf, &Config)
	if err != nil {
		log.Fatalf("Configuration file format error: %v", err)
	}
}
