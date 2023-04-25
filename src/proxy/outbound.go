package proxy

import (
	"errors"
	"go_proxy/transport"
	"log"
	"net"
)

type Outbound struct {
	Address  string
	Protocol string
	Transmit string
	WsPath   string
}

func (outbound Outbound) Dial(targetAddr string, payload []byte) (out OutboundConnect, err error) {
	if outbound.Protocol == "btp" {
		// *BtpOutbound implemented OutboundConnect,
		// here we return the pointer of BtpOutbound, which is an OutboundConnect
		// simply, *BtpOutbound is OutboundConnect
		var conn, err = transport.Dial(
			outbound.Address,
			outbound.Transmit,
			outbound.WsPath,
		)
		if err != nil {
			log.Println("btp failed to dial to", outbound.Address, ", because", err)
			return nil, err
		}
		log.Println("btp connect to", outbound.Address)
		out = &BtpOutbound{conn: conn}
	} else {
		var conn, err = transport.Dial(
			targetAddr,
			outbound.Transmit,
			outbound.WsPath,
		)
		if err != nil {
			log.Println("free failed to dial to", outbound.Address, ", because", err)
			return nil, err
		}
		log.Println("free connect to", targetAddr)
		out = &FreeOutbound{conn: conn}
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

type FreeOutbound struct {
	conn net.Conn
}

func (outbound *FreeOutbound) Connect(targetAddr string, payload []byte) (err error) {
	// pointer receiver: just implement the method of *FreeOutbound
	_ = targetAddr
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
	conn net.Conn
}

func (outbound *BtpOutbound) Connect(targetAddr string, payload []byte) (err error) {
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
