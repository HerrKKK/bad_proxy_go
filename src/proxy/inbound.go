package proxy

import (
	"fmt"
	"go_proxy/protocols"
)

func HttpAccept(proxy *Proxy) error {
	proxy.buffer = make([]byte, 8196)
	_, err := proxy.Inbound.Read(proxy.buffer[:])
	if err != nil {
		return err
	}
	if proxy.buffer == nil {
		fmt.Println("buffer null after first http parse")
	}
	request, err := protocols.HTTPRequest{}.Parse(proxy.buffer[:])
	if err != nil {
		return err
	}
	proxy.targetAddr = request.Address
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		proxy.Inbound.Write([]byte(response))
		fmt.Println("https connect")
		proxy.buffer = make([]byte, 8196) // clear
		length, _ := proxy.Inbound.Read(proxy.buffer)
		proxy.buffer = proxy.buffer[:length]
	}
	return nil
}

func BtpAccept(proxy *Proxy) error {
	proxy.buffer = make([]byte, 8196)
	length, err := proxy.Inbound.Read(proxy.buffer)
	if err != nil {
		return err
	}
	request, err := protocols.BTPRequest{}.Parse(proxy.buffer[:length])
	if err != nil {
		return err
	}
	proxy.targetAddr = request.Address
	//copy(proxy.buffer, request.Payload)
	proxy.buffer = request.Payload // trilling extra zero
	return nil
}
