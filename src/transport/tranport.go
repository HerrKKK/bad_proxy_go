package transport

import (
	"crypto/tls"
	"go_proxy/structure"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
)

func Dial(address string, transmit string) (conn net.Conn, err error) {
	if transmit == "tls" {
		return tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	} else if transmit == "ws" {
		return websocket.Dial(
			"ws://localhost:9999/bp",
			"",
			"http://localhost/",
		)
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
	} else if transmit == "ws" {
		listener := &WsListener{
			queue: *structure.NewQueue(),
			ch:    make(chan net.Conn),
		}
		http.Handle("/bp", websocket.Handler(listener.handle))
		go http.ListenAndServe(":9999", nil)
		return listener, nil
	}
	return net.Listen("tcp", address)
}
