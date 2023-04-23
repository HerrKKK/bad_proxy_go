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
		_, err := proxy.Inbound.Read(proxy.buffer[:])
		if err != nil {
			fmt.Println("second read failure")
			return err
		}
	}
	return nil
}

func BtpAccept(proxy *Proxy) error {
	_, err := proxy.Inbound.Read(proxy.buffer[:])
	if err != nil {
		return err
	}
	request, err := protocols.BTPRequest{}.Parse(proxy.buffer[:])
	if err != nil {
		return err
	}
	proxy.targetAddr = request.Address
	//fmt.Println("outbound target addr is", request.Address)
	copy(proxy.buffer, request.Payload)
	//fmt.Println("recv btp buffer is")
	//fmt.Println(string(proxy.buffer[:]))
	return nil
}
