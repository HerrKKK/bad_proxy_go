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

	mainProxy := newProxy(config)
	mainProxy.Start()
	quit := make(chan os.Signal)
	<-quit
}

func newProxy(config Config) (newProxy proxy.Proxy) {
	for _, in := range config.Inbound {
		newInbound := proxy.Inbound{
			Address:     in.Host + ":" + in.Port,
			Protocol:    in.Protocol,
			Transmit:    in.Transmit,
			WsPath:      in.WsPath,
			TlsCertPath: in.TlsCertPath,
			TlsKeyPath:  in.TlsKeyPath,
		}
		newProxy.Inbound = append(newProxy.Inbound, newInbound)
	}

	for _, out := range config.Outbound {
		newOutbound := proxy.Outbound{
			Address:  out.Host + ":" + out.Port,
			Protocol: out.Protocol,
			Transmit: out.Transmit,
			WsPath:   out.WsPath,
		}
		newProxy.Outbound = append(newProxy.Outbound, newOutbound)
	}
	return
}
