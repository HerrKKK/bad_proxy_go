package transport

import (
	"net"
)

func Dial(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

func Listen(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
