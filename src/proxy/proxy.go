package proxy

import (
	"fmt"
	"io"
)

type Proxy struct {
	Inbound  Inbound
	Outbound Outbound
}

type RoutingPackage struct {
	Address string
	Payload []byte
}

func (proxy Proxy) Proxy() {
	for {
		fmt.Println("loop")
		in, _ := proxy.Inbound.Accept()
		go proxy.process(in)
	}
}

func (proxy Proxy) process(in InboundConnect) {
	routingPackage, err := in.Connect() // handshake
	if err != nil {
		return
	}
	// routing to find outbound template
	out, err := proxy.Outbound.Dial() // handshake
	if err != nil || out == nil {
		fmt.Println("outbound dial failure")
		return
	}
	err = out.Connect(routingPackage)
	if err != nil {
		return
	}
	go io.Copy(in, out)
	io.Copy(out, in)
	in.Close()
	out.Close()
}
