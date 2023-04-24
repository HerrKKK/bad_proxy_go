package proxy

import (
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
		in, _ := proxy.Inbound.Accept()
		go proxy.process(in)
	}
}

func (proxy Proxy) process(in InboundConnect) {
	address, payload, err := in.Connect() // handshake
	if err != nil {
		return
	}
	// routing to find outbound template
	out, err := proxy.Outbound.Dial(address, payload) // handshake
	if err != nil {
		return
	}
	go io.Copy(in, out)
	io.Copy(out, in)

	in.Close()
	out.Close()
}
