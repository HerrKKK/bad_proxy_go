package structure

import (
	"golang.org/x/net/websocket"
	"log"
	"sync"
)

type ConnectQueue struct {
	data []**websocket.Conn
	cond *sync.Cond
}

func NewQueue() *ConnectQueue {
	queue := ConnectQueue{
		data: make([]**websocket.Conn, 0),
		cond: sync.NewCond(new(sync.Mutex)),
	}
	return &queue
}

func (queue *ConnectQueue) Push(conn **websocket.Conn) {
	queue.cond.L.Lock()
	queue.data = append(queue.data, conn)
	log.Println("first is ", queue.data[0])
	queue.cond.Signal()
	queue.cond.L.Unlock()
}

func (queue *ConnectQueue) Pop() (front **websocket.Conn) {
	queue.cond.L.Lock()
	for len(queue.data) == 0 {
		queue.cond.Wait() // unlock cond.L and pending for signal and re-lock on waken
	}
	front = queue.data[0]
	queue.data = queue.data[1:]
	queue.cond.L.Unlock()
	return front
}

func (queue *ConnectQueue) Size() int {
	return len(queue.data)
}
