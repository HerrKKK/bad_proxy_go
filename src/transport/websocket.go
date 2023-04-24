package transport

import (
	"go_proxy/structure"
	"golang.org/x/net/websocket"
	"net"
)

type WsListener struct {
	queue structure.ConnectQueue
	addr  net.Addr
}

func (listener *WsListener) Accept() (net.Conn, error) {
	return listener.queue.Pop(), nil // Pop wait for producer
}

func (listener *WsListener) Close() error {
	return nil
}

func (listener *WsListener) Addr() net.Addr {
	return listener.addr
}

func (listener *WsListener) handle(conn *websocket.Conn) {
	// hack websocket connection to our queue
	go listener.queue.Push(conn)
}
