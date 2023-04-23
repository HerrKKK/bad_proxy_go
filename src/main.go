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
			Address: config.Outbound.Host + ":" + config.Outbound.Port,
		}
		if config.Inbound.Protocol == "http" {
			//fmt.Println("inbound http")
			instance.Accept = proxy.HttpAccept
		} else if config.Inbound.Protocol == "btp" {
			//fmt.Println("inbound btp")
			instance.Accept = proxy.BtpAccept
		}
		if config.Outbound.Protocol == "btp" {
			//fmt.Println("outbound btp")
			instance.Dial = proxy.BtpDial
		} else {
			//fmt.Println("outbound free")
			instance.Dial = proxy.FreeDial
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
