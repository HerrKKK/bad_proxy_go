package main

import (
	"fmt"
	"go_proxy/proxy"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Panic(err)
	}
	for {
		instance := proxy.Proxy{
			Accept: proxy.HttpConnect,
			Dial:   proxy.FreeConnect,
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
