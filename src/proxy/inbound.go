package proxy

import (
	"fmt"
	"go_proxy/protocols"
)

func HttpConnect(proxy *Proxy) error {
	_, err := proxy.Inbound.Read(proxy.buffer[:])
	if err != nil {
		return err
	}
	request, err := protocols.HTTPRequest{}.Parse(proxy.buffer[:])
	if err != nil {
		return err
	}
	if request.Method == "CONNECT" {
		proxy.Inbound.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		proxy.buffer = nil
		fmt.Println("https connect")
	}
	proxy.targetAddr = request.Address
	return nil
}
