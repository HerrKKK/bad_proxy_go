package transport

import (
	"crypto/tls"
	"log"
	"net"
)

var plain = false

func Dial(address string, transmit string) (conn net.Conn, err error) {
	if transmit == "tls" {
		return tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	}
	return net.Dial("tcp", address)
}

func Listen(address string, transmit string) (listener net.Listener, err error) {
	if transmit == "tls" {
		cert, err := tls.LoadX509KeyPair(
			"certs/localhost_certificate.pem",
			"certs/localhost_key.pem",
		)
		if err != nil {
			log.Println(err)
			return listener, err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", address, config)
	}
	return net.Listen("tcp", address)
}
