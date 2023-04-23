package proxy

import (
	"io"
	"net"
)

type Proxy struct {
	Inbound    net.Conn
	Outbound   net.Conn
	Accept     func(proxy *Proxy) error
	Dial       func(proxy *Proxy) error
	Connect    func(proxy *Proxy) error
	buffer     []byte
	targetAddr string // parsed from request format: google.com:443
	Address    string // specified by config
}

func (proxy Proxy) Proxy() {
	proxy.buffer = make([]byte, 8196)
	err := proxy.Accept(&proxy) // client connection
	if err != nil {
		return
	}
	err = proxy.Dial(&proxy) // 4L connection
	if err != nil {
		return
	}

	go io.Copy(proxy.Outbound, proxy.Inbound)
	io.Copy(proxy.Inbound, proxy.Outbound)

	proxy.Inbound.Close()
	proxy.Outbound.Close()
}
