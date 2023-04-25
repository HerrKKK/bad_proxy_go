package transport

import (
	"go_proxy/structure"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"time"
)

type WsListener struct {
	queue structure.ConnectQueue
	addr  net.Addr
	ch    chan net.Conn
}

func (listener *WsListener) Accept() (conn net.Conn, err error) {
	conn = <-listener.ch
	buffer := make([]byte, 8196)
	length, _ := conn.Read(buffer)
	log.Println("accept len is", length)
	return
}

func (listener *WsListener) Close() error {
	return nil
}

func (listener *WsListener) Addr() net.Addr {
	return listener.addr
}

func (listener *WsListener) handle(conn *websocket.Conn) {
	ws := WsConnect{conn: &conn}
	log.Println("ws recv addr is", &conn)
	buffer := make([]byte, 8196)
	length, _ := conn.Read(buffer)
	log.Println("handle len is", length)
	listener.ch <- ws
}

type WsConnect struct {
	conn **websocket.Conn
}

func (ws WsConnect) Read(b []byte) (n int, err error) {
	log.Println("ws core addr is", ws.conn)
	return (*ws.conn).Read(b)
}

func (ws WsConnect) Write(b []byte) (n int, err error) {
	return (*ws.conn).Write(b)
}

func (ws WsConnect) Close() error {
	return (*ws.conn).Close()
}

func (ws WsConnect) LocalAddr() net.Addr {
	return (*ws.conn).LocalAddr()
}

func (ws WsConnect) RemoteAddr() net.Addr {
	return (*ws.conn).RemoteAddr()
}

func (ws WsConnect) SetDeadline(t time.Time) error {
	return (*ws.conn).SetDeadline(t)
}

func (ws WsConnect) SetReadDeadline(t time.Time) error {
	return (*ws.conn).SetReadDeadline(t)
}

func (ws WsConnect) SetWriteDeadline(t time.Time) error {
	return (*ws.conn).SetWriteDeadline(t)
}
