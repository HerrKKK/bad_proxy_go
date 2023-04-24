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
	Transmit string
}

func (outbound Outbound) Dial(targetAddr string, payload []byte) (out OutboundConnect, err error) {
	err = nil
	if outbound.Protocol == "btp" {
		// *BtpOutbound implemented OutboundConnect,
		// here we return the pointer of BtpOutbound, which is an OutboundConnect
		// simply, *BtpOutbound is OutboundConnect
		out = &BtpOutbound{address: outbound.Address}
	} else {
		out = &FreeOutbound{address: targetAddr}
	}
	//outbound.conn, err = transport.Dial(out.address, transmit)
	out.Connect(targetAddr, payload, outbound.Transmit)
	return
}

type OutboundConnect interface {
	Connect(targetAddr string, payload []byte, transmit string) error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type FreeOutbound struct {
	conn    net.Conn
	address string
}

func (outbound *FreeOutbound) Connect(targetAddr string, payload []byte, transmit string) (err error) {
	// pointer receiver: just implement the method of *FreeOutbound
	outbound.conn, err = transport.Dial(targetAddr, transmit)
	if err != nil || payload == nil {
		fmt.Println("free failed to connect, addr is ", outbound.address)
		return err
	}
	_, err = outbound.conn.Write(payload)
	return err
}

func (outbound *FreeOutbound) Read(b []byte) (int, error) {
	return outbound.conn.Read(b)
}

func (outbound *FreeOutbound) Write(b []byte) (int, error) {
	return outbound.conn.Write(b)
}

func (outbound *FreeOutbound) Close() error {
	if outbound.conn == nil {
		return nil
	}
	return outbound.conn.Close()
}

type BtpOutbound struct {
	conn    net.Conn
	address string
}

func (outbound *BtpOutbound) Connect(targetAddr string, payload []byte, transmit string) (err error) {
	outbound.conn, err = transport.Dial(outbound.address, transmit)
	if err != nil || payload == nil {
		fmt.Println("btp failed to connect, addr is ", outbound.address)
		return
	}

	fmt.Println("btp connect to", outbound.address)
	payload = append(
		[]byte{uint8(len(targetAddr))}, // must less than 255 for uint8
		append([]byte(targetAddr), payload[:]...)...,
	)
	_, err = outbound.conn.Write(payload)
	return err
}

func (outbound *BtpOutbound) Read(b []byte) (int, error) {
	if outbound.conn == nil {
		return 0, errors.New("nil conn")
	}
	return outbound.conn.Read(b)
}

func (outbound *BtpOutbound) Write(b []byte) (int, error) {
	if outbound.conn == nil {
		return 0, errors.New("nil conn")
	}
	return outbound.conn.Write(b)
}

func (outbound *BtpOutbound) Close() error {
	if outbound.conn == nil {
		return nil
	}
	return outbound.conn.Close()
}
