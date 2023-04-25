package proxy

import (
	"go_proxy/protocols"
	"go_proxy/transport"
	"net"
)

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
	Listener    net.Listener
	Secret      string
	Protocol    string
	Address     string
	Transmit    transport.ProtocolType
	WsPath      string
	TlsCertPath string
	TlsKeyPath  string
}

func (inbound *Inbound) Listen() (err error) {
	inbound.Listener, err = transport.Listen(
		inbound.Address,
		inbound.Transmit,
		inbound.WsPath,
		inbound.TlsCertPath,
		inbound.TlsKeyPath,
	)
	return
}

func (inbound *Inbound) Accept() (inConn InboundConnect, err error) {
	conn, err := inbound.Listener.Accept()
	if err != nil {
		return
	}
	if inbound.Protocol == "http" {
		inConn = HttpInbound{conn: conn}
	} else if inbound.Protocol == "btp" {
		inConn = BtpInbound{conn: conn, secret: inbound.Secret}
	}
	return
}

type InboundConnect interface {
	Connect() (string, []byte, error)
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type HttpInbound struct {
	conn net.Conn
}

func (inbound HttpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196)
	length, err := inbound.conn.Read(payload[:])
	if err != nil {
		return
	}
	request, err := protocols.HTTPRequest{}.Parse(payload[:])
	if err != nil {
		return
	}
	targetAddr = request.Address
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		_, err = inbound.conn.Write([]byte(response))
		if err != nil {
			return
		}
		payload = make([]byte, 8196) // clear
		length, err = inbound.conn.Read(payload)
		if err != nil {
			return
		}
	}
	payload = payload[:length]
	return
}

func (inbound HttpInbound) Read(b []byte) (int, error) {
	return inbound.conn.Read(b)
}

func (inbound HttpInbound) Write(b []byte) (int, error) {
	return inbound.conn.Write(b)
}

func (inbound HttpInbound) Close() error {
	return inbound.conn.Close()
}

type BtpInbound struct {
	conn   net.Conn
	secret string
}

func (inbound BtpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196)
	length, err := inbound.conn.Read(payload)
	if err != nil {
		return
	}
	request, err := protocols.ParseBtpRequest(payload[:length])
	if err != nil {
		return
	}
	err = request.Validate(inbound.secret)
	if err != nil {
		return
	}
	return request.Address, request.Payload, nil
}

func (inbound BtpInbound) Read(b []byte) (int, error) {
	return inbound.conn.Read(b)
}

func (inbound BtpInbound) Write(b []byte) (int, error) {
	return inbound.conn.Write(b)
}

func (inbound BtpInbound) Close() error {
	return inbound.conn.Close()
}
