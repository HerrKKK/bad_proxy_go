package proxy

import (
	"fmt"
	"io"
	"net"
)

type Proxy struct {
	Inbound    net.Conn
	Outbound   net.Conn
	Accept     func(proxy *Proxy) error
	Dial       func(proxy *Proxy) error
	buffer     []byte
	targetAddr string
}

func (proxy Proxy) Proxy() {
	proxy.buffer = make([]byte, 1024)
	err := proxy.Accept(&proxy)
	if err != nil {
		return
	}
	err = proxy.Dial(&proxy)
	if err != nil {
		return
	}

	fmt.Println("connect to " + proxy.targetAddr)

	go io.Copy(proxy.Outbound, proxy.Inbound)
	io.Copy(proxy.Inbound, proxy.Outbound)

	proxy.Inbound.Close()
	proxy.Outbound.Close()
}
