package proxy

import (
	"go_proxy/protocols"
	"go_proxy/transport"
	"log"
)

type OutboundConfig struct {
	Tag      string `json:"tag"`
	Secret   string `json:"secret"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Transmit string `json:"transmit"`
	WsPath   string `json:"ws_path"`
}

type Outbound struct {
	tag      string
	secret   string
	address  string
	protocol string
	transmit transport.ProtocolType
	wsPath   string
}

func (outbound *Outbound) Dial(targetAddr string, payload []byte) (out OutboundConnect, err error) {
	switch outbound.protocol {
	// *BtpOutbound implemented OutboundConnect,
	// here we return the pointer of BtpOutbound, which is an OutboundConnect
	// simply, *BtpOutbound is OutboundConnect
	case "btp":
		var conn, err = transport.Dial(
			outbound.address,
			outbound.transmit,
			outbound.wsPath,
		)
		if err != nil {
			return nil, err
		}
		log.Println("btp connect to", outbound.address)
		out = &protocols.BtpOutbound{Conn: conn, Secret: outbound.secret}
	case "socks":
		var conn, err = transport.Dial(outbound.address, transport.TCP, "")
		if err != nil {
			return nil, err
		}
		log.Println("socks connect to", outbound.address)
		out = &protocols.Socks5Outbound{Conn: conn}
	default: // free
		var conn, err = transport.Dial(
			targetAddr,
			outbound.transmit,
			outbound.wsPath,
		)
		if err != nil {
			return nil, err
		}
		log.Println("free connect to", targetAddr)
		out = &protocols.FreeOutbound{Conn: conn}
	}
	err = out.Connect(targetAddr, payload)
	return
}

type OutboundConnect interface {
	Connect(targetAddr string, payload []byte) error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}
