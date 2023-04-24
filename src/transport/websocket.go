package transport

import (
	"net"
	"time"
)

type TCPListener struct {
	listener net.Listener
}

func (listener TCPListener) Accept() (net.Conn, error) {
	return listener.listener.Accept()
}

func (listener TCPListener) Close() error {
	return listener.listener.Close()
}

func (listener TCPListener) Addr() net.Addr {
	return listener.listener.Addr()
}

func TCPListen(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	return TCPListener{listener: listener}, err
}

type TCPConnect struct {
	conn net.Conn
}

func (tcp TCPConnect) Read(b []byte) (int, error) {
	return tcp.conn.Read(b)
}

func (tcp TCPConnect) Write(b []byte) (int, error) {
	return tcp.conn.Write(b)
}

func (tcp TCPConnect) Close() error {
	return tcp.conn.Close()
}

func (tcp TCPConnect) LocalAddr() net.Addr {
	return tcp.conn.LocalAddr()
}

func (tcp TCPConnect) RemoteAddr() net.Addr {
	return tcp.conn.RemoteAddr()
}

func (tcp TCPConnect) SetDeadline(t time.Time) error {
	return tcp.conn.SetDeadline(t)
}

func (tcp TCPConnect) SetReadDeadline(t time.Time) error {
	return tcp.conn.SetReadDeadline(t)
}

func (tcp TCPConnect) SetWriteDeadline(t time.Time) error {
	return tcp.conn.SetWriteDeadline(t)
}
