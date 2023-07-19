package protocols

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	VERSION = 0x05

	MethodNoAuth = 0x00

	CmdConnect = 0x01

	AtypIpv4       = 0x01
	AtypDomainname = 0x03
	AtypIpv6       = 0x04

	RepSuccess = 0x00

	Ipv4Length = 4
	Ipv6Length = 16
)

type Socks5Message struct {
	version     uint8
	command     uint8 // REP in response
	rsv         uint8
	addressType uint8
	host        string
	port        int
}

func (socks *Socks5Message) toByteArray() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, socks.version) // socks5
	// 0x01: CONNECT with tcp, 0x02: BIND, waiting for, 0x03: UDP ASSOCIATE
	_ = binary.Write(bytesBuffer, binary.BigEndian, socks.command)
	_ = binary.Write(bytesBuffer, binary.BigEndian, socks.rsv)
	_ = binary.Write(bytesBuffer, binary.BigEndian, socks.addressType)
	if socks.addressType == AtypDomainname {
		_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(len([]byte(socks.host))))
	}
	_ = binary.Write(bytesBuffer, binary.BigEndian, []byte(socks.host))
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint16(socks.port))
	return bytesBuffer.Bytes()
}

func (socks *Socks5Message) Print() {
	log.Printf("version: %d\n", socks.version)
	log.Printf("command: %d\n", socks.command)
	log.Printf("addressType: %d\n", socks.addressType)
	log.Printf("host: %s\n", socks.host)
	log.Printf("port: %d\n", socks.port)
}

func encodeSockS5Request(targetAddr string) (request Socks5Message) {
	request.version = VERSION
	request.command = CmdConnect
	hnp := strings.Split(targetAddr, ":")
	request.host = hnp[0]
	request.port, _ = strconv.Atoi(hnp[1])

	ip := net.ParseIP(request.host)
	if ip == nil {
		request.addressType = AtypDomainname // domain name
	} else if strings.Contains(ip.String(), ":") {
		request.addressType = AtypIpv6 // ipv4
	} else {
		request.addressType = AtypIpv4 // ipv6
	}
	return
}

func parseSockS5Message(data []byte) (response Socks5Message, err error) {
	if len(data) < 4 {
		err = errors.New("too short message")
		return
	}
	response.version = data[0]
	response.command = data[1]
	response.rsv = data[2]
	response.addressType = data[3]
	pos := 4
	switch response.addressType {
	case AtypIpv4: // curl --socks5-hostname
		response.host = string(data[pos : pos+Ipv4Length])
		pos += Ipv4Length
	case AtypDomainname:
		domainLength := int(data[pos])
		pos++
		if pos+domainLength > len(data)-2 {
			err = errors.New("failed to parse domain")
			return
		}
		response.host = string(data[pos : pos+domainLength])
		pos += domainLength
	case AtypIpv6:
		response.host = string(data[pos : pos+Ipv6Length])
		pos += Ipv6Length
	default:
		err = errors.New("wrong address type")
		return
	}
	response.port = int(data[pos])<<8 + int(data[pos+1])
	return
}

type Socks5Outbound struct {
	Conn     net.Conn
	username string
	password string
}

func (outbound *Socks5Outbound) Connect(targetAddr string, payload []byte) (err error) {
	// socks5 version, length of methods, methods: no-auth only
	buffer := []byte{VERSION, 0x01, MethodNoAuth}
	if _, err = outbound.Conn.Write(buffer); err != nil {
		return
	}
	buffer = make([]byte, 1024)
	_, err = outbound.Conn.Read(buffer)          // the chosen encryption
	if err != nil || buffer[2] != MethodNoAuth { // only no encryption supported
		return
	}

	request := encodeSockS5Request(targetAddr)
	if _, err = outbound.Conn.Write(request.toByteArray()); err != nil {
		return
	}

	buffer = make([]byte, 1024)
	if _, err = outbound.Conn.Read(buffer); err != nil {
		return
	}
	response, err := parseSockS5Message(buffer)
	if err != nil || response.command != RepSuccess {
		return
	}

	_, err = outbound.Conn.Write(payload)
	return
}

func (outbound *Socks5Outbound) Read(b []byte) (int, error) {
	return outbound.Conn.Read(b)
}

func (outbound *Socks5Outbound) Write(b []byte) (int, error) {
	return outbound.Conn.Write(b)
}

func (outbound *Socks5Outbound) Close() error {
	return outbound.Conn.Close()
}

type Socks5Inbound struct {
	Conn     net.Conn
	Host     string
	Port     int
	username string
	password string
}

func (inbound *Socks5Inbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 1024) // return rawdata on error
	if _, err = inbound.Conn.Read(payload); err != nil {
		return
	}

	if payload[0] != VERSION {
		log.Println("wrong socks version")
		return
	}

	if _, err = inbound.Conn.Write([]byte{VERSION, MethodNoAuth}); err != nil {
		return
	}

	payload = make([]byte, 1024)
	if _, err = inbound.Conn.Read(payload); err != nil {
		return
	}

	request, err := parseSockS5Message(payload)
	if err != nil || request.command != CmdConnect {
		return // only CMD_CONNECTION is supported now
	}

	response := Socks5Message{
		version:     VERSION,
		command:     RepSuccess,
		rsv:         0x00,
		addressType: AtypIpv4, // just bind to 0.0.0.0
		host:        string([]byte{0x00, 0x00, 0x00, 0x00}),
		port:        inbound.Port,
	} // BND.ADDR/BND.PORT: the real relay server address

	/*
		Response should be sent after tcp connected according to socks5 protocol
		while we just tell the client we succeeded and receive the first message
		because we relay data to outbound instead of a tcp connection.
	*/
	if _, err = inbound.Conn.Write(response.toByteArray()); err != nil {
		return
	}

	payload = make([]byte, 8196)
	if _, err = inbound.Conn.Read(payload); err != nil {
		return
	}

	return request.host + ":" + strconv.Itoa(request.port), payload, err
}

func (inbound *Socks5Inbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	_ = reverseLocalAddr
	_ = rawdata
}

func (inbound *Socks5Inbound) Read(b []byte) (int, error) {
	return inbound.Conn.Read(b)
}

func (inbound *Socks5Inbound) Write(b []byte) (int, error) {
	return inbound.Conn.Write(b)
}

func (inbound *Socks5Inbound) Close() error {
	return inbound.Conn.Close()
}
