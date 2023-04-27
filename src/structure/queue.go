package structure

import (
	"errors"
	"sync"
)

type Queue[T string | *ACAutomaton] struct {
	data    []T
	head    int
	tail    int
	maxSize int
	lock    *sync.Mutex
}

func NewQueue[T string | *ACAutomaton](maxSize int, initSize int) (queue *Queue[T]) {
	queue = &Queue[T]{
		data:    make([]T, initSize),
		maxSize: maxSize,
		head:    0,
		tail:    0,
		lock:    new(sync.Mutex),
	}
	return
}

func (queue *Queue[T]) Push(value T) (err error) { // from back
	queue.lock.Lock()
	queue.data[queue.tail] = value
	capacity := len(queue.data)
	queue.tail = (queue.tail + 1) % capacity
	if queue.tail == queue.head {
		err = queue.expand()
	}
	queue.lock.Unlock()
	return
}

func (queue *Queue[T]) Pop() (res T) { // from front
	queue.lock.Lock()
	defer queue.lock.Unlock()

	if queue.size() == 0 {
		panic("no element in the list should never happen")
	}
	res = queue.data[queue.head]
	capacity := len(queue.data)
	queue.head = (capacity + queue.head - 1) % capacity
	return
}

func (queue *Queue[T]) Size() int {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	return queue.size()
}

func (queue *Queue[T]) size() int { // concurrent INSECURE
	capacity := len(queue.data)
	return (capacity + queue.tail - queue.head) % capacity
}

func (queue *Queue[T]) expand() (err error) {
	capacity := len(queue.data)
	if queue.maxSize > 0 && capacity == queue.maxSize {
		return errors.New("reach max size")
	}
	newCapacity := capacity * 2
	if queue.maxSize > 0 && newCapacity > queue.maxSize {
		newCapacity = queue.maxSize
	}
	size := queue.size()
	data := make([]T, newCapacity)
	if queue.head < queue.tail {
		copy(data, queue.data[queue.head:queue.tail])
	} else {
		copy(data, queue.data[queue.head:])
		copy(data[capacity-queue.head:], queue.data[:queue.tail])
	}
	queue.head = 0
	queue.tail = size
	queue.data = data
	return
}
