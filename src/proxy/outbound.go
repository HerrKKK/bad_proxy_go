package proxy

import (
	"fmt"
	"net"
)

func FreeDial(proxy *Proxy) error {
	fmt.Println("outbound free connect to", proxy.targetAddr)
	outbound, err := net.Dial("tcp", proxy.targetAddr)
	if err != nil {
		fmt.Println("free outbound connect failure, target is ", proxy.targetAddr)
		return err
	}
	proxy.Outbound = outbound

	if proxy.buffer == nil {
		fmt.Println("free dial buffer is nil")
		return nil
	}
	_, err = proxy.Outbound.Write(proxy.buffer[:])
	return err
}

func BtpDial(proxy *Proxy) error {
	//fmt.Println("outbound btp connect to", proxy.Address)
	outbound, err := net.Dial("tcp", proxy.Address)
	if err != nil {
		fmt.Println("btp outbound connect failure, target is ", proxy.Address)
		return err
	}
	proxy.Outbound = outbound

	if proxy.buffer == nil {
		fmt.Println("btp dial buffer is nil")
		return nil
	}
	//fmt.Println("inbound target len is ", []byte{uint8(len(proxy.targetAddr))})
	payload := append(
		[]byte{uint8(len(proxy.targetAddr))}, // must less than 255 for uint8
		append([]byte(proxy.targetAddr), proxy.buffer[:]...)...,
	)
	//fmt.Println("sent btp payload is")
	//fmt.Println(string(proxy.buffer[:]))
	_, err = proxy.Outbound.Write(payload)
	return err
}
