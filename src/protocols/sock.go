package protocols

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	VERSION = 0x05

	METHOD_NO_AUTH   = 0x00
	METHOD_GSSAPI    = 0x01
	METHOD_USR_PWD   = 0x02
	METHOD_NO_METHOD = 0xFF

	CMD_CONNECT       = 0x01
	CMD_BIND          = 0x02
	CMD_UDP_ASSOCIATE = 0x03

	ATYP_IPV4       = 0x01
	ATYP_DOMAINNAME = 0x03
	ATYP_IPV6       = 0x04

	REP_SUCCESS             = 0x00
	REP_FAILURE             = 0x01
	REP_NOT_ALLOWED         = 0x02
	REP_NETWORK_UNREACHABLE = 0x03
	REP_HOST_UNREACHABLE    = 0x04
	REP_CONNECTION_REFUSED  = 0x05
	REP_TTL_EXPIRED         = 0x06
	REP_CMD_UNSUPPORTED     = 0x07
	REP_ATYP_UNSUPPORTED    = 0x08
)

type SockS5Package struct {
	version     uint8
	command     uint8 // REP in response
	rsv         uint8
	addressType uint8
	host        string
	port        int
}

func (sock *SockS5Package) toByteArray() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, sock.version) // socks5
	// 0x01: CONNECT with tcp, 0x02: BIND, waiting for, 0x03: UDP ASSOCIATE
	_ = binary.Write(bytesBuffer, binary.BigEndian, sock.command)
	_ = binary.Write(bytesBuffer, binary.BigEndian, sock.rsv)
	_ = binary.Write(bytesBuffer, binary.BigEndian, sock.addressType)
	if sock.addressType == 0x03 { // write a byte to represent length of host
		_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(len([]byte(sock.host))))
	}
	_ = binary.Write(bytesBuffer, binary.BigEndian, []byte(sock.host))
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint16(sock.port))
	return bytesBuffer.Bytes()
}

func (sock *SockS5Package) Print() {
	log.Printf("version: %d\n", sock.version)
	log.Printf("command: %d\n", sock.command)
	log.Printf("addressType: %d\n", sock.addressType)
	log.Printf("host: %s\n", sock.host)
	log.Printf("port: %d\n", sock.port)
}

func encodeSockS5Request(targetAddr string) (request SockS5Package) {
	request.version = VERSION
	request.command = CMD_CONNECT
	hnp := strings.Split(targetAddr, ":")
	request.host = hnp[0]
	request.port, _ = strconv.Atoi(hnp[1])

	ip := net.ParseIP(request.host)
	if ip == nil {
		request.addressType = ATYP_DOMAINNAME // domain name
	} else if strings.Contains(ip.String(), ":") {
		request.addressType = ATYP_IPV6 // ipv4
	} else {
		request.addressType = ATYP_IPV4 // ipv6
	}
	return
}

func parseSockS5Response(data []byte, length int) (response SockS5Package) {
	if length < 8 {
		return
	}
	response.version = data[0]
	response.command = data[1]
	response.rsv = data[2]
	response.addressType = data[3]
	response.host = string(data[4 : length-2])
	response.port = int(binary.BigEndian.Uint32(data[length-2:]))
	return
}

type SockS5Outbound struct {
	Conn     net.Conn
	username string
	password string
}

func (outbound *SockS5Outbound) Connect(targetAddr string, payload []byte) (err error) {
	// socks5 version, length of methods, methods: no-auth only
	buffer := []byte{VERSION, 0x01, METHOD_NO_AUTH}
	_, err = outbound.Conn.Write(buffer)
	if err != nil {
		return
	}
	buffer = make([]byte, 1024)
	_, err = outbound.Conn.Read(buffer)            // the chosen encryption
	if err != nil || buffer[2] != METHOD_NO_AUTH { // only no encryption supported
		return
	}

	request := encodeSockS5Request(targetAddr)
	if _, err = outbound.Conn.Write(request.toByteArray()); err != nil {
		return
	}

	buffer = make([]byte, 1024)
	length, err := outbound.Conn.Read(buffer) // read response
	response := parseSockS5Response(buffer, length)
	if response.command != REP_SUCCESS {
		return
	}

	if len(payload) != 0 {
		_, err = outbound.Conn.Write(payload)
	}
	return
}

func (outbound *SockS5Outbound) Read(b []byte) (int, error) {
	return outbound.Conn.Read(b)
}

func (outbound *SockS5Outbound) Write(b []byte) (int, error) {
	return outbound.Conn.Write(b)
}

func (outbound *SockS5Outbound) Close() error {
	return outbound.Conn.Close()
}
