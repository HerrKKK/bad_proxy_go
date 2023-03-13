package proxy

import (
	"fmt"
	"net"
)

func FreeConnect(proxy *Proxy) error {
	outbound, err := net.Dial("tcp", proxy.targetAddr)
	if err != nil {
		fmt.Println("outbound connect failure")
		return err
	}
	proxy.Outbound = outbound

	if proxy.buffer == nil {
		return nil
	}
	_, err = proxy.Outbound.Write(proxy.buffer[:])
	return err
}
