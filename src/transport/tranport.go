package transport

import (
	"crypto/tls"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
)

func Dial(address string, transmit string, wsPath string) (conn net.Conn, err error) {
	if transmit == "tls" {
		return tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	} else if transmit == "ws" || transmit == "wss" {
		return websocket.Dial(
			transmit+"://"+address+wsPath,
			"",
			"http://localhost/",
		)
	}
	return net.Dial("tcp", address)
}

func Listen(
	address string,
	transmit string,
	wsPath string,
	tlsCertPath string,
	tlsKeyPath string,
) (listener net.Listener, err error) {
	if transmit == "tls" {
		cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			log.Println(err)
			return listener, err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", address, config)
	} else if transmit == "ws" {
		listener := &WsListener{ch: make(chan net.Conn)}
		http.Handle(wsPath, websocket.Handler(listener.handle))
		go http.ListenAndServe(address, nil)
		return listener, nil
	}
	return net.Listen("tcp", address) // plain tcp
}
