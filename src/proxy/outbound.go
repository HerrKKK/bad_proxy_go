package proxy

import (
	"fmt"
	"net"
)

func FreeTCPDial(proxy *Proxy) error {
	outbound, err := net.Dial("tcp", proxy.targetAddr)
	if err != nil {
		fmt.Println("outbound connect failure")
		return err
	}
	proxy.Outbound = outbound
	return nil
}

func FixTCPDial(proxy *Proxy) error {
	outbound, err := net.Dial("tcp", proxy.Address)
	if err != nil {
		fmt.Println("outbound connect failure")
		return err
	}
	proxy.Outbound = outbound
	return nil
}

func HTTPConnect(proxy *Proxy) error {
	if proxy.buffer == nil {
		return nil
	}
	_, err := proxy.Outbound.Write(proxy.buffer[:])
	return err
}
