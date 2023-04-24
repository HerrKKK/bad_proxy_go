package proxy

import (
	"fmt"
	"go_proxy/protocols"
	"net"
)

type Inbound struct {
	Listener net.Listener
	Protocol string
}

func (inbound Inbound) Accept() (InboundConnect, error) {
	conn, _ := inbound.Listener.Accept()
	if inbound.Protocol == "http" {
		return HttpInbound{socket: conn}, nil
	} else if inbound.Protocol == "btp" {
		return BtpInbound{socket: conn}, nil
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
	socket net.Conn
}

func (inbound HttpInbound) Connect() (string, []byte, error) {
	buffer := make([]byte, 8196)
	length, err := inbound.socket.Read(buffer[:])
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
		_, err = inbound.socket.Write([]byte(response))
		if err != nil {
			return "", nil, err
		}
		fmt.Println("https connect")
		buffer = make([]byte, 8196) // clear
		length, _ = inbound.socket.Read(buffer)
		fmt.Println("http", request.Address, "recv length is", length)
	}
	buffer = buffer[:length]
	return targetAddr, buffer, nil
}

func (inbound HttpInbound) Read(b []byte) (int, error) {
	return inbound.socket.Read(b)
}

func (inbound HttpInbound) Write(b []byte) (int, error) {
	return inbound.socket.Write(b)
}

func (inbound HttpInbound) Close() error {
	return inbound.socket.Close()
}

type BtpInbound struct {
	socket net.Conn
}

func (inbound BtpInbound) Connect() (string, []byte, error) {
	buffer := make([]byte, 8196)
	length, err := inbound.socket.Read(buffer)
	if err != nil {
		return "", nil, err
	}
	request, err := protocols.BTPRequest{}.Parse(buffer[:length])
	if err != nil {
		return "", nil, err
	}
	return request.Address, request.Payload, nil
}

func (inbound BtpInbound) Read(b []byte) (int, error) {
	return inbound.socket.Read(b)
}

func (inbound BtpInbound) Write(b []byte) (int, error) {
	return inbound.socket.Write(b)
}

func (inbound BtpInbound) Close() error {
	return inbound.socket.Close()
}
