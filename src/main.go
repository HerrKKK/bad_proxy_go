package main

import (
	"fmt"
	"go_proxy/proxy"
	"log"
	"net"
)

func main() {
	config, err := ReadConfig("conf/server_config.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("config: " + config.Inbound.Protocol)

	l, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Panic(err)
	}
	for {
		instance := proxy.Proxy{
			Accept:  proxy.HttpConnect,
			Dial:    proxy.FreeTCPDial,
			Connect: proxy.HTTPConnect,
		}
		fmt.Println("listen on 10000")
		conn, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		instance.Inbound = conn
		go instance.Proxy()
	}
}
