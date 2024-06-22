package transport

import (
	"crypto/tls"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
)

type ProtocolType string

const (
	TCP ProtocolType = "tcp"
	TLS ProtocolType = "tls"
	WS  ProtocolType = "ws"
	WSS ProtocolType = "wss"
)

func (protocol ProtocolType) Str() string {
	return string(protocol)
}

func GetProtocol(protocol string) ProtocolType {
	return ProtocolType(protocol)
}

func Dial(protocol ProtocolType, address string) (conn net.Conn, err error) {
	switch protocol {
	case TLS:
		return tls.Dial(TCP.Str(), address, &tls.Config{})
	case WS, WSS:
		return websocket.Dial(protocol.Str()+"://"+address, "", "http://localhost/")
	default:
		return net.Dial(TCP.Str(), address)
	}
}

func Listen(
	address string,
	protocol ProtocolType,
	wsPath string,
	tlsCertPath string,
	tlsKeyPath string,
) (listener net.Listener, err error) {
	switch protocol {
	case TLS:
		cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			return listener, err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen(TCP.Str(), address, config)
	case WS:
		listener := &WsListener{ch: make(chan net.Conn)}
		http.Handle(wsPath, websocket.Handler(listener.handle))
		go func() {
			err := http.ListenAndServe(address, nil)
			if err != nil {
				panic(err)
			}
		}()
		return listener, nil
	default:
		return net.Listen(TCP.Str(), address)
	}
}
