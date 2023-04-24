package main

import (
	"flag"
	"fmt"
	"go_proxy/proxy"
	"go_proxy/transport"
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
	listener, _ := transport.Listen(config.Inbound.Host + ":" + config.Inbound.Port)
	proxy := proxy.Proxy{
		Inbound: proxy.Inbound{
			Listener: listener,
			Protocol: config.Inbound.Protocol,
		},
		Outbound: proxy.Outbound{
			Address:  config.Outbound.Host + ":" + config.Outbound.Port,
			Protocol: config.Outbound.Protocol,
		},
	}
	proxy.Proxy()
}
