package config

import (
	"fmt"
	"os"
	"strings"
)

var defaultConfFile = "./server.conf"

var confMap = make(map[string]string)

var consulConfMap = make(map[string]string)

func SetConfFile(file string) {
	defaultConfFile = file
}

func GetConf(key string) string {
	return confMap[key]
}

func GetConsulConfMap(key string) string {
	return consulConfMap[key]
}

func Init() {
	conf, err := os.ReadFile(defaultConfFile)
	if err != nil {
		fmt.Println("read config file error , use default config!")
		return
	}
	confStr := string(conf)
	confs := strings.Split(confStr, "\n")
	for i := 0; i < len(confs); i++ {
		confs[i] = strings.Trim(confs[i], "\r")
		if strings.HasPrefix(confs[i], "//") || confs[i] == "" {
			continue
		}
		oneConf := strings.Split(confs[i], "=")
		if len(oneConf) == 2 {
			confMap[strings.Trim(oneConf[0], "\r")] = strings.Trim(oneConf[1], "\r")
		}
	}
	fmt.Println("load  config file success!")
}
