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
	Connect() (RoutingPackage, error)
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type HttpInbound struct {
	socket net.Conn
}

func (inbound HttpInbound) Connect() (RoutingPackage, error) {
	buffer := make([]byte, 8196)
	_, err := inbound.socket.Read(buffer[:])
	if err != nil {
		return RoutingPackage{}, err
	}
	if buffer == nil {
		fmt.Println("buffer null after first http parse")
		return RoutingPackage{}, nil
	}
	request, err := protocols.HTTPRequest{}.Parse(buffer[:])
	if err != nil {
		return RoutingPackage{}, err
	}
	targetAddr := request.Address
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		inbound.socket.Write([]byte(response))
		fmt.Println("https connect")
		buffer = make([]byte, 8196) // clear
		length, _ := inbound.socket.Read(buffer)
		fmt.Println("http", request.Address, "recv length is", length)
		buffer = buffer[:length]
	}
	return RoutingPackage{Address: targetAddr, Payload: buffer}, nil
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

func (inbound BtpInbound) Connect() (RoutingPackage, error) {
	buffer := make([]byte, 8196)
	length, err := inbound.socket.Read(buffer)
	if err != nil {
		return RoutingPackage{}, err
	}
	request, err := protocols.BTPRequest{}.Parse(buffer[:length])
	if err != nil {
		return RoutingPackage{}, err
	}
	return RoutingPackage{Address: request.Address, Payload: request.Payload}, nil
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
