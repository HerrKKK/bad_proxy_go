package transport

import (
	"crypto/tls"
	"log"
	"net"
)

var plain = false

func Dial(address string, protocol string) (conn net.Conn, err error) {
	if protocol == "freedom" {
		return net.Dial("tcp", address)
	}
	return tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
}

func Listen(address string, protocol string) (listener net.Listener, err error) {
	if protocol == "http" {
		return net.Listen("tcp", address)
	}
	cert, err := tls.LoadX509KeyPair(
		"certs/localhost_certificate.pem",
		"certs/localhost_key.pem",
	)
	if err != nil {
		log.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", address, config)
}
