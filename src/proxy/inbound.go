package proxy

import (
	"go_proxy/protocols"
	"go_proxy/transport"
	"net"
	"strconv"
	"strings"
)

type FallbackConfig struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
}

type InboundConfig struct {
	Secret      string `json:"secret"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Transmit    string `json:"transmit"`
	WsPath      string `json:"ws_path"`
	TlsCertPath string `json:"tls_cert_path"`
	TlsKeyPath  string `json:"tls_key_path"`
}

type Inbound struct {
	listener    net.Listener
	secret      string
	protocol    string
	address     string
	transmit    transport.ProtocolType
	wsPath      string
	tlsCertPath string
	tlsKeyPath  string
}

func (inbound *Inbound) Listen() (err error) {
	inbound.listener, err = transport.Listen(
		inbound.address,
		inbound.transmit,
		inbound.wsPath,
		inbound.tlsCertPath,
		inbound.tlsKeyPath,
	)
	return
}

func (inbound *Inbound) Accept() (inConn InboundConnect, err error) {
	conn, err := inbound.listener.Accept()
	if err != nil {
		return
	}

	switch inbound.protocol {
	case "http":
		inConn = &protocols.HttpInbound{Conn: conn}
	case "btp":
		inConn = &protocols.BtpInbound{Conn: conn, Secret: inbound.secret}
	case "socks":
		hnp := strings.Split(inbound.address, ":")
		port, _ := strconv.Atoi(hnp[1])
		inConn = &protocols.Socks5Inbound{
			Conn: conn,
			Host: hnp[0],
			Port: port,
		}
	}
	return
}

type InboundConnect interface {
	Connect() (string, []byte, error)
	Fallback(reverseLocalAddr string, rawdata []byte)
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}
