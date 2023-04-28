package proxy

import (
	"go_proxy/router"
	"go_proxy/transport"
	"io"
	"log"
	"strings"
)

type Proxy struct {
	Inbounds  []Inbound
	Outbounds map[string]*Outbound
	Fallback  FallbackConfig
	routers   []router.Router
}

type Config struct {
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
	Router    []router.Config  `json:"routers"`
	Fallback  FallbackConfig   `json:"fallback"`
}

func NewProxy(config Config) (newProxy Proxy) {
	newProxy.Fallback = config.Fallback
	newProxy.Outbounds = make(map[string]*Outbound, 10)
	newProxy.routers = make([]router.Router, 0)
	for _, r := range config.Router {
		rules := make([]string, 0)
		for _, rule := range r.Rules {
			rules = append(rules, rule)
		}
		newRouter, err := router.NewRouter(r.Tag, rules)
		if err != nil {
			log.Println("wrong router:", r.Tag)
			continue
		}
		newProxy.routers = append(newProxy.routers, *newRouter)
	}

	for _, in := range config.Inbounds {
		newInbound := Inbound{
			Secret:      in.Secret,
			Address:     in.Host + ":" + in.Port,
			Protocol:    in.Protocol,
			Transmit:    transport.GetProtocol(in.Transmit),
			WsPath:      in.WsPath,
			TlsCertPath: in.TlsCertPath,
			TlsKeyPath:  in.TlsKeyPath,
		}
		newProxy.Inbounds = append(newProxy.Inbounds, newInbound)
	}

	for _, out := range config.Outbounds {
		newOutbound := Outbound{
			Tag:      out.Tag,
			Secret:   out.Secret,
			Address:  out.Host + ":" + out.Port,
			Protocol: out.Protocol,
			Transmit: transport.GetProtocol(out.Transmit),
			WsPath:   out.WsPath,
		}
		_, exist := newProxy.Outbounds[out.Tag]
		if exist == true {
			log.Fatalln("duplicate tag")
		}
		newProxy.Outbounds[out.Tag] = &newOutbound
	}
	return
}

func (proxy Proxy) Start() {
	if proxy.Fallback.LocalAddr != "" && proxy.Fallback.RemoteAddr != "" {
		go StartReverseProxy(proxy.Fallback.LocalAddr, proxy.Fallback.RemoteAddr)
	}
	for _, inbound := range proxy.Inbounds {
		go func(inbound Inbound) {
			log.Println("listen on", inbound.Address)
			err := inbound.Listen()
			if err != nil {
				log.Fatalf("inbound on %s dead!\n", inbound.Address)
				return
			}
			for {
				in, err := inbound.Accept()
				if err != nil {
					log.Printf("inbound on %s failed to accept!\n", inbound.Address)
					continue
				}
				go proxy.proxy(in)
			}
		}(inbound)
	}
}

func (proxy Proxy) proxy(in InboundConnect) {
	address, payload, err := in.Connect() // handshake
	defer in.Close()
	if err != nil {
		log.Println("inbound connect failed:", err)
		in.Fallback(proxy.Fallback.LocalAddr, payload)
		return
	}
	// routing to find outbound template
	outbound := proxy.route(address)
	out, err := outbound.Dial(address, payload) // handshake
	defer out.Close()
	if err != nil {
		log.Printf("outbound dial to %s failed\n", outbound.Address)
		return
	}
	go io.Copy(in, out)
	_, _ = io.Copy(out, in)
}

func (proxy Proxy) route(address string) (outbound Outbound) {
	// If s does not contain sep and sep is not empty,
	// Split returns a slice of length 1 whose only element is s
	tag := ""
	host := strings.Split(address, ":")[0]
	for _, r := range proxy.routers {
		if r.MatchAny(host) == true {
			tag = r.Tag
		}
	}
	return *proxy.Outbounds[tag]
}
