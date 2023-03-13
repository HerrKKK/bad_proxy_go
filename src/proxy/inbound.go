package proxy

import (
	"fmt"
	"go_proxy/protocols"
)

func HttpAccept(proxy *Proxy) error {
	_, err := proxy.Inbound.Read(proxy.buffer[:])
	if err != nil {
		return err
	}
	request, err := protocols.HTTPRequest{}.Parse(proxy.buffer[:])
	if err != nil {
		return err
	}
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		proxy.Inbound.Write([]byte(response))
		proxy.buffer = nil
		fmt.Println("https connect")
	}
	proxy.targetAddr = request.Address
	return nil
}
