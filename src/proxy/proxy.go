package proxy

import (
	"fmt"
	"io"
	"time"
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
	defer in.Close()
	fmt.Println("process start")
	t := time.Now()
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
	defer out.Close()
	err = out.Connect(routingPackage)
	if err != nil {
		return
	}
	go io.Copy(in, out)
	io.Copy(out, in)

	fmt.Print("process end ")
	fmt.Println(time.Since(t) / time.Millisecond)
}
