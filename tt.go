package main

import (
	"sync"
)

type ConnectQueue struct {
	cond *sync.Cond
	size int
}

func NewQueue() *ConnectQueue {
	queue := ConnectQueue{
		cond: sync.NewCond(new(sync.Mutex)),
		size: 10,
	}
	return &queue
}

func (queue *ConnectQueue) Push() {
	for {
		queue.cond.L.Lock()
		queue.size++
		queue.cond.L.Unlock()
		queue.cond.Signal()
	}
}

func (queue *ConnectQueue) Pop() {
	for {
		queue.cond.L.Lock()
		for queue.size == 0 {
			queue.cond.Wait() // unlock cond.L and pending for signal and re-lock on waken
		}
		queue.size--
		queue.cond.L.Unlock()
	}
}

func main() {
	queue := NewQueue()
	//rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		go queue.Push()
	}
	for i := 0; i < 5; i++ {
		go queue.Pop()
	}

	var quit chan string
	<-quit
}
