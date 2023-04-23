package transport

import (
	"net"
)

type TCPNet struct {
	listener net.Listener
}

func (tcp TCPNet) Listen(address string) (Network, error) {
	listener, err := net.Listen("tcp", address)
	tcp.listener = listener
	return tcp, err
}

func (tcp TCPNet) Accept() (Transport, error) {
	return tcp.listener.Accept()
}

func (tcp TCPNet) Dial(address string) (Transport, error) {
	return net.Dial("tcp", address)
}

type TCPConn struct {
	conn net.Conn
}

func (tcp TCPConn) Read(buffer []byte) (int, error) {
	return tcp.conn.Read(buffer)
}

func (tcp TCPConn) Write(buffer []byte) (int, error) {
	return tcp.conn.Write(buffer)
}

func (tcp TCPConn) Close() error {
	return tcp.conn.Close()
}
