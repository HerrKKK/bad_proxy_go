package main

import (
	"flag"
	"go_proxy/proxy"
	"log"
	"os"
)

func main() {
	var configPath string
	flag.StringVar(
		&configPath,
		"c",
		"conf/server_config.json",
		"config file path",
	)
	flag.Parse()
	config, err := ReadConfig(configPath)
	if err != nil {
		log.Println(err)
		return
	}
	mainProxy := proxy.NewProxy(config)
	mainProxy.Start()
	<-make(chan os.Signal)
}
