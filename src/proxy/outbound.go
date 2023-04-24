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

func (outbound Outbound) Dial(address string, payload []byte) (out OutboundConnect, err error) {
	err = nil
	if outbound.Protocol == "btp" {
		// *BtpOutbound implemented OutboundConnect,
		// here we return the pointer of BtpOutbound, which is an OutboundConnect
		// simply, *BtpOutbound is OutboundConnect
		out = &BtpOutbound{address: outbound.Address}
	} else {
		out = &FreeOutbound{address: outbound.Address}
	}
	out.Connect(address, payload)
	return
}

type OutboundConnect interface {
	Connect(address string, payload []byte) error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type FreeOutbound struct {
	socket  net.Conn
	address string
}

func (outbound *FreeOutbound) Connect(address string, payload []byte) (err error) {
	// pointer receiver: just implement the method of *FreeOutbound
	outbound.socket, err = transport.Dial(address, "freedom")
	if err != nil || payload == nil {
		fmt.Println("free failed to connect, addr is ", outbound.address)
		return err
	}
	_, err = outbound.socket.Write(payload)
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

func (outbound *BtpOutbound) Connect(address string, payload []byte) (err error) {
	outbound.socket, err = transport.Dial(outbound.address, "btp")
	if err != nil || payload == nil {
		fmt.Println("btp failed to connect, addr is ", outbound.address)
		return
	}

	fmt.Println("btp connect to", outbound.address)
	payload = append(
		[]byte{uint8(len(address))}, // must less than 255 for uint8
		append([]byte(address), payload[:]...)...,
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
