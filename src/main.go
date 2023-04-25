package main

import (
	"flag"
	"fmt"
	"go_proxy/proxy"
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
		fmt.Println(err)
		return
	}

	mainProxy := proxy.NewProxy(config)
	mainProxy.Start()
	quit := make(chan os.Signal)
	<-quit
}
