package protocols

import (
	"errors"
	"sync"
)

type StringQueue struct {
	data    []string
	head    int
	tail    int
	maxSize int
	lock    *sync.Mutex
}

func newStringQueue(maxSize int, initSize int) (queue *StringQueue) {
	queue = &StringQueue{
		data:    make([]string, initSize),
		maxSize: maxSize,
		head:    0,
		tail:    0,
		lock:    new(sync.Mutex),
	}
	return
}

func (queue *StringQueue) Push(value string) (err error) { // from back
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

func (queue *StringQueue) Pop() (res string) { // from front
	queue.lock.Lock()
	defer queue.lock.Unlock()

	if queue.size() == 0 {
		panic("should never happen")
	}
	res = queue.data[queue.head]
	capacity := len(queue.data)
	queue.head = (capacity + queue.head - 1) % capacity
	return
}

func (queue *StringQueue) size() int { // concurrent INSECURE
	capacity := len(queue.data)
	return (capacity + queue.tail - queue.head) % capacity
}

func (queue *StringQueue) expand() (err error) {
	capacity := len(queue.data)
	if capacity == queue.maxSize {
		return errors.New("reach max size")
	}
	newCapacity := capacity * 2
	if newCapacity > queue.maxSize {
		newCapacity = queue.maxSize
	}
	size := queue.size()
	data := make([]string, newCapacity)
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

type BtpLRU struct {
	data    map[string]bool
	queue   *StringQueue
	lock    *sync.Mutex
	maxSize int
}

func (lru *BtpLRU) init(maxSize int) {
	lru.maxSize = maxSize
	lru.data = make(map[string]bool, lru.maxSize)
	lru.queue = newStringQueue(lru.maxSize, 2)
	lru.lock = new(sync.Mutex)
}

func (lru *BtpLRU) Add(key string) (err error) {
	lru.lock.Lock()
	defer lru.lock.Unlock()
	_, exist := lru.data[key]
	if exist == true {
		return errors.New("exising digest, possible replay attack")
	}

	if lru.queue.size() == lru.maxSize {
		delete(lru.data, lru.queue.Pop())
	}

	lru.data[key] = true
	return lru.queue.Push(key)
}

var lru *BtpLRU
var once sync.Once

func GetBtpCache() *BtpLRU {
	once.Do(func() {
		instance := &BtpLRU{}
		instance.init(210000)
		lru = instance
	})
	return lru
}
