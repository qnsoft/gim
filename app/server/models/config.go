/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         config.go
@ Create Time:  2019-07-29 14:55
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
type IM struct {
	Public `json:"im"`
}

// Push Server
type Push struct {
	Public `json:"push"`
}

// Services
type Services struct {
	IM
	Push
}

// config struct
type Conf struct {
	Redis    Redis
	Services Services
}

// Json formatted printing method
func (c Conf) Print() {
	if buf, err := json.MarshalIndent(c, "", "\t"); err != nil {
		log.Println("Json format error: ", err)
	} else {
		log.Println(string(buf))
	}
}

// loading config
func init() {
	if buf, err := ioutil.ReadFile("app/server/configs/config.json"); err != nil {
		log.Fatalf("Unable to load configuration file: %v", err)
	} else {
		if err = json.Unmarshal(buf, &Config); err != nil {
			log.Fatalf("Configuration file format error: %v", err)
		}
	}
}
