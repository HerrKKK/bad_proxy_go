package proxy

import (
	"go_proxy/protocols"
	"go_proxy/transport"
	"io"
	"net"
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
	if inbound.protocol == "http" {
		inConn = &HttpInbound{conn: conn}
	} else if inbound.protocol == "btp" {
		inConn = &BtpInbound{conn: conn, secret: inbound.secret}
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

type HttpInbound struct {
	conn net.Conn
}

func (inbound *HttpInbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	_ = reverseLocalAddr
	_ = rawdata
	return
}

func (inbound *HttpInbound) Connect() (targetAddr string, payload []byte, err error) {
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

func (inbound *HttpInbound) Read(b []byte) (int, error) {
	return inbound.conn.Read(b)
}

func (inbound *HttpInbound) Write(b []byte) (int, error) {
	return inbound.conn.Write(b)
}

func (inbound *HttpInbound) Close() error {
	return inbound.conn.Close()
}

type BtpInbound struct {
	conn   net.Conn
	secret string
}

func (inbound *BtpInbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	out, err := net.Dial("tcp", reverseLocalAddr)
	defer inbound.Close()
	defer out.Close()
	if err != nil {
		return
	}
	_, _ = out.Write(rawdata)

	go io.Copy(inbound, out)
	_, _ = io.Copy(out, inbound)
	return
}

func (inbound *BtpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196) // return rawdata on error
	length, err := inbound.conn.Read(payload)
	if err != nil {
		return
	}
	request, err := protocols.ParseBtpRequest(payload[:length])
	if err != nil {
		return
	}
	err = request.Validate(inbound.secret)
	if err != nil { // try to handle http connection
		return
	}
	return request.Address, request.Payload, nil
}

func (inbound *BtpInbound) Read(b []byte) (int, error) {
	return inbound.conn.Read(b)
}

func (inbound *BtpInbound) Write(b []byte) (int, error) {
	return inbound.conn.Write(b)
}

func (inbound *BtpInbound) Close() error {
	return inbound.conn.Close()
}
