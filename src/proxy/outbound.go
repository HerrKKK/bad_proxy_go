package proxy

import (
	"errors"
	"fmt"
	"go_proxy/transport"
	"net"
)

type Outbound struct {
	Address  string
	Protocol string
}

func (outbound Outbound) Dial() (OutboundConnect, error) {
	if outbound.Protocol == "btp" {
		// *BtpOutbound implemented OutboundConnect,
		// here we return the pointer of BtpOutbound, which is an OutboundConnect
		// simply, *BtpOutbound is OutboundConnect
		return &BtpOutbound{address: outbound.Address}, nil
	}
	return &FreeOutbound{address: outbound.Address}, nil
}

type OutboundConnect interface {
	Connect(routingPackage RoutingPackage) error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type FreeOutbound struct {
	socket  net.Conn
	address string
}

func (outbound *FreeOutbound) Connect(routingPackage RoutingPackage) (err error) {
	// pointer receiver: just implement the method of *FreeOutbound
	outbound.socket, err = transport.Dial(routingPackage.Address)
	if err != nil {
		fmt.Println("free failed to connect, addr is ", outbound.address)
		return err
	}

	if routingPackage.Payload == nil {
		fmt.Println("free dial buffer is nil")
		return nil
	}
	fmt.Println("free connect to",
		routingPackage.Address,
		"length is",
		len(routingPackage.Payload))
	_, err = outbound.socket.Write(routingPackage.Payload)
	return err
}

func (outbound *FreeOutbound) Read(b []byte) (int, error) {
	return outbound.socket.Read(b)
}

func (outbound *FreeOutbound) Write(b []byte) (int, error) {
	return outbound.socket.Write(b)
}

func (outbound *FreeOutbound) Close() error {
	if outbound.socket == nil {
		return nil
	}
	return outbound.socket.Close()
}

type BtpOutbound struct {
	socket  net.Conn
	address string
}

func (outbound *BtpOutbound) Connect(routingPackage RoutingPackage) error {
	socket, err := transport.Dial(outbound.address)
	if err != nil || socket == nil {
		fmt.Println("btp failed to connect, addr is ", outbound.address)
		return err
	}
	outbound.socket = socket

	if routingPackage.Payload == nil {
		fmt.Println("btp dial buffer is nil")
		return nil
	}
	fmt.Println("btp connect to", outbound.address)
	payload := append(
		[]byte{uint8(len(routingPackage.Address))}, // must less than 255 for uint8
		append([]byte(routingPackage.Address), routingPackage.Payload[:]...)...,
	)
	_, err = outbound.socket.Write(payload)
	return err
}

func (outbound *BtpOutbound) Read(b []byte) (int, error) {
	if outbound.socket == nil {
		return 0, errors.New("nil socket")
	}
	return outbound.socket.Read(b)
}

func (outbound *BtpOutbound) Write(b []byte) (int, error) {
	if outbound.socket == nil {
		return 0, errors.New("nil socket")
	}
	return outbound.socket.Write(b)
}

func (outbound *BtpOutbound) Close() error {
	if outbound.socket == nil {
		return nil
	}
	return outbound.socket.Close()
}
