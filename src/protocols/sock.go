package protocols

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
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
	request.version = 0x05
	request.command = 0x01
	hnp := strings.Split(targetAddr, ":")
	request.host = hnp[0]
	request.port, _ = strconv.Atoi(hnp[1])

	request.addressType = 0x03 // domain name
	if net.ParseIP(request.host) != nil {
		request.addressType = 0x01 // ipv4, v6 not supported
	}
	return
}

func parseSockS5Response(data []byte, length int) (response SockS5Package) {
	if length < 8 {
		log.Printf("socks response length is %d\n", length)
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
	// socks5 version, length of methods, methods: only no encrypted
	buffer := []byte{0x05, 0x01, 0x00}
	_, err = outbound.Conn.Write(buffer)
	if err != nil {
		return
	}
	buffer = make([]byte, 1024)
	_, err = outbound.Conn.Read(buffer)  // the chosen encryption
	if err != nil || buffer[2] != 0x00 { // only no encryption supported
		return
	}

	request := encodeSockS5Request(targetAddr)
	log.Println(request.toByteArray())
	_, err = outbound.Conn.Write(request.toByteArray())
	if err != nil {
		log.Printf("failed to write for %s\n", err.Error())
		return
	}

	buffer = make([]byte, 1024)
	log.Printf("before read")
	length, err := outbound.Conn.Read(buffer) // read response
	log.Printf("after read")
	response := parseSockS5Response(buffer, length)
	if response.command != 0x00 {
		log.Printf("response failed %d\n", response.command)
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
