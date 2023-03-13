package main

import (
	"flag"
	"fmt"
	"go_proxy/proxy"
	"log"
	"net"
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
	l, err := net.Listen("tcp", config.Inbound.Host+":"+config.Inbound.Port)
	if err != nil {
		log.Panic(err)
	}
	for {
		instance := proxy.Proxy{
			Accept: proxy.HttpAccept,
			Dial:   proxy.FreeConnect,
		}
		fmt.Println("listen on " + config.Inbound.Host + ":" + config.Inbound.Port)
		conn, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		instance.Inbound = conn
		go instance.Proxy()
	}
}
