/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         main.go
@ Create Time:  2019-07-18 12:38
@ Software:     GoLand
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gim/src/im"
	"gim/src/models"
	. "gim/src/routers"
	"io/ioutil"
	"log"
	"os"
)

var (
	help   bool
	post   string
	imHost string
	imPort string
)

type Config struct {
	Redis models.Redis
}

func (conf Config) Load(path string) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(buf, conf)
}

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.StringVar(&post, "port", "8080", "GIM restful api port")
	flag.StringVar(&imHost, "imhost", "0.0.0.0", "GIM server host")
	flag.StringVar(&imPort, "import", "8088", "GIM server port")
	flag.Parse()

	_ = os.Setenv("PORT", post)
	_ = os.Setenv("IMHost", imHost)
	_ = os.Setenv("IMPort", imPort)

	Conf := Config{}
	Conf.Load("src/config/config.json")
	fmt.Printf("Config: %+v\n", Conf)

	if help {
		flag.Usage()
	} else {
		// Start GIM server
		im.Run()
		// Start Restful API
		if err := App.Run(); err != nil {
			log.Fatalln("Service startup failed !", err)
		}
	}
}
