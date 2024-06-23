package proxy

import (
	"C"
	"go_proxy/router"
	"go_proxy/transport"
	"io"
	"log"
	"strings"
)

type Proxy struct {
	inbounds  []Inbound
	outbounds map[string]*Outbound
	routers   []router.Router
}

type Config struct {
	Inbounds   []InboundConfig  `json:"inbounds"`
	Outbounds  []OutboundConfig `json:"outbounds"`
	Router     []router.Config  `json:"routers"`
	RouterPath string
}

const (
	BTP   = "btp"
	SOCKS = "socks"
	HTTP  = "http"
)

func NewProxy(config Config) (newProxy Proxy) {
	newProxy.outbounds = make(map[string]*Outbound, 10)
	newProxy.routers = make([]router.Router, 0)
	for _, r := range config.Router {
		rules := make([]string, 0)
		for _, rule := range r.Rules {
			rules = append(rules, rule)
		}
		newRouter, err := router.NewRouter(r.Tag, rules, config.RouterPath)
		if err != nil {
			log.Println("wrong router:", r.Tag)
			continue
		}
		newProxy.routers = append(newProxy.routers, *newRouter)
	}

	for _, in := range config.Inbounds {
		newInbound := Inbound{
			secret:      in.Secret,
			address:     in.Host + ":" + in.Port,
			protocol:    in.Protocol,
			transmit:    transport.GetProtocol(in.Transmit),
			wsPath:      in.WsPath,
			tlsCertPath: in.TlsCertPath,
			tlsKeyPath:  in.TlsKeyPath,
		}
		newProxy.inbounds = append(newProxy.inbounds, newInbound)
	}

	for _, out := range config.Outbounds {
		newOutbound := Outbound{
			tag:      out.Tag,
			secret:   out.Secret,
			address:  out.Host + ":" + out.Port,
			protocol: out.Protocol,
			transmit: transport.GetProtocol(out.Transmit),
			wsPath:   out.WsPath,
		}

		if _, exist := newProxy.outbounds[out.Tag]; exist == true {
			log.Fatalln("duplicate outbound tag")
		}
		newProxy.outbounds[out.Tag] = &newOutbound
	}
	return
}

func (proxy Proxy) Start() {
	for _, inbound := range proxy.inbounds {
		go func(inbound Inbound) { // TODO: Why reference pass does not work.
			log.Println(inbound.protocol, "listen on", inbound.address)
			err := inbound.Listen()
			if err != nil {
				log.Fatalf("inbound on %s dead, %s\n", inbound.address, err.Error())
				return
			}
			for {
				in, err := inbound.Accept()
				if err != nil {
					log.Printf("inbound on %s failed to accept!\n", inbound.address)
					continue
				}
				go proxy.proxy(in)
			}
		}(inbound)
	}
}

func (proxy Proxy) proxy(in InboundConnect) {
	defer func() { // recover any panic to avoid quiting from main loop.
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	address, payload, err := in.Connect() // handshake
	if err != nil {
		log.Println("inbound connect failed:", err)
		in.Fallback(payload)
		return
	}
	// routing to find outbound template
	outbound := proxy.route(address)
	out, err := outbound.Dial(address, payload) // handshake
	if err != nil {
		log.Printf("outbound dial to %s failed\n", outbound.address)
		return
	}
	go func() {
		if _, err := io.Copy(in, out); err != nil {
			log.Printf("write to %s failed\n", outbound.address)
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		log.Printf("read from %s failed\n", outbound.address)
	}
	_ = in.Close()
	_ = out.Close()
}

func (proxy Proxy) route(address string) Outbound {
	// If s does not contain sep and sep is not empty,
	// Split returns a slice of length 1 whose only element is s
	tag := ""
	host := strings.Split(address, ":")[0]
	for _, r := range proxy.routers {
		if r.MatchAny(host) == true {
			tag = r.Tag
			break
		}
	}
	outbound, exist := proxy.outbounds[tag]
	if exist == false {
		return *proxy.outbounds[""]
	}
	return *outbound
}
