package proxy

import (
	"io"
	"log"
)

type Proxy struct {
	Inbounds  []Inbound
	Outbounds []Outbound
}

type Config struct {
	Inbounds  []InboundConfig  `json:"inbound"`
	Outbounds []OutboundConfig `json:"outbound"`
}

func NewProxy(config Config) (newProxy Proxy) {
	for _, in := range config.Inbounds {
		newInbound := Inbound{
			Secret:      in.Secret,
			Address:     in.Host + ":" + in.Port,
			Protocol:    in.Protocol,
			Transmit:    in.Transmit,
			WsPath:      in.WsPath,
			TlsCertPath: in.TlsCertPath,
			TlsKeyPath:  in.TlsKeyPath,
		}
		newProxy.Inbounds = append(newProxy.Inbounds, newInbound)
	}

	for _, out := range config.Outbounds {
		newOutbound := Outbound{
			Secret:   out.Secret,
			Address:  out.Host + ":" + out.Port,
			Protocol: out.Protocol,
			Transmit: out.Transmit,
			WsPath:   out.WsPath,
		}
		newProxy.Outbounds = append(newProxy.Outbounds, newOutbound)
	}
	return
}

func (proxy Proxy) Start() {
	for _, inbound := range proxy.Inbounds {
		go func(inbound Inbound) {
			log.Println("listen on", inbound.Address)
			err := inbound.Listen()
			if err != nil {
				log.Fatalf(
					"inbound on %s dead!\n",
					inbound.Address,
				)
				return
			}
			for {
				in, err := inbound.Accept()
				if err != nil {
					log.Printf(
						"inbound on %s failed to accept!\n",
						inbound.Address,
					)
					continue
				}
				go proxy.proxy(in)
			}
		}(inbound)
	}
}

func (proxy Proxy) proxy(in InboundConnect) {
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
	return proxy.Outbounds[0]
}
