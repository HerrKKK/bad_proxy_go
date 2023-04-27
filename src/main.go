package main

import (
	"flag"
	"go_proxy/proxy"
	"go_proxy/structure"
	"log"
	"os"
)

var patterns = [256]string{
	"he",
	"she",
	"his",
	"her",
}

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
	structure.Test()
	mainProxy := proxy.NewProxy(config)
	mainProxy.Start()
	<-make(chan os.Signal)
}
