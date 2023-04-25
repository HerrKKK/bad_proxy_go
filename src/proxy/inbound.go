package proxy

import (
	"go_proxy/protocols"
	"go_proxy/transport"
	"log"
	"net"
)

type InboundConfig struct {
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
	Protocol    string
	Address     string
	Transmit    string
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

func (inbound *Inbound) Accept() (InboundConnect, error) {
	conn, _ := inbound.Listener.Accept()
	if inbound.Protocol == "http" {
		return HttpInbound{conn: conn}, nil
	} else if inbound.Protocol == "btp" {
		return BtpInbound{conn: conn}, nil
	}
	return nil, nil
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

func (inbound HttpInbound) Connect() (string, []byte, error) {
	buffer := make([]byte, 8196)
	length, err := inbound.conn.Read(buffer[:])
	if err != nil || buffer == nil {
		return "", nil, err
	}
	request, err := protocols.HTTPRequest{}.Parse(buffer[:])
	if err != nil {
		return "", nil, err
	}
	targetAddr := request.Address
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		_, err = inbound.conn.Write([]byte(response))
		if err != nil {
			return "", nil, err
		}
		buffer = make([]byte, 8196) // clear
		length, _ = inbound.conn.Read(buffer)
	}
	buffer = buffer[:length]
	log.Println(
		"http connect from",
		request.Address,
		"to",
		inbound.conn.RemoteAddr(),
	)
	return targetAddr, buffer, nil
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
	conn net.Conn
}

func (inbound BtpInbound) Connect() (string, []byte, error) {
	buffer := make([]byte, 8196)
	length, err := inbound.conn.Read(buffer)
	if err != nil {
		return "", nil, err
	}
	request, err := protocols.BTPRequest{}.Parse(buffer[:length])
	if err != nil {
		return "", nil, err
	}
	log.Println(
		"btp connect from",
		request.Address,
		"to",
		inbound.conn.RemoteAddr(),
	)
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
