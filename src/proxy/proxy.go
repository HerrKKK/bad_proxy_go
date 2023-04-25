package proxy

import (
	"io"
	"log"
)

type Proxy struct {
	Inbound  []Inbound
	Outbound []Outbound
}

func (proxy Proxy) Start() {
	for index, _ := range proxy.Inbound {
		go func(index int) {
			log.Println("listen on", proxy.Inbound[index].Address)
			_ = proxy.Inbound[index].Listen()
			for {
				in, _ := proxy.Inbound[index].Accept()
				go proxy.process(in)
			}
		}(index)
	}
}

func (proxy Proxy) process(in InboundConnect) {
	address, payload, err := in.Connect() // handshake
	if err != nil {
		return
	}
	// routing to find outbound template
	outbound := proxy.route(address)
	out, err := outbound.Dial(address, payload) // handshake
	if err != nil {
		return
	}
	go io.Copy(in, out)
	io.Copy(out, in)

	in.Close()
	out.Close()
}

func (proxy Proxy) route(address string) (outbound Outbound) {
	_ = address
	return proxy.Outbound[0]
}
