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
)

type Config struct {
	Redis Redis
}

func (conf Config) Load(path string) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(buf, conf)
}
