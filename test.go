package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cond sync.Cond //创建全局条件变量

func producter(count *int) {
	for {
		cond.L.Lock() //go进程进来后，不多比比，先上锁
		//for *count == 10 { /*注意这里是for而不是if，判断缓冲
		//	区里是不是有太多数据，如果满了，则持续阻塞，直至对面发来signal信号*/
		//	cond.Wait()
		//}
		*count = *count + 1
		fmt.Println("已生产", *count)
		cond.L.Unlock() //生产完毕后解锁并给对面发信号
		cond.Signal()
		time.Sleep(time.Millisecond * 300)
	}
}
func consumer(count *int) {
	for {
		cond.L.Lock() //以下都是跟上面一样的
		for *count == 0 {
			cond.Wait()
		}
		*count = *count - 1
		fmt.Println("已消费：", *count)
		cond.L.Unlock()
		//cond.Signal()
		time.Sleep(time.Millisecond * 300)
	}
}
func main() {
	//给全局变量加上一个锁的功能，相当于给他加个装备
	cond.L = new(sync.Mutex)
	count := 10
	var quit chan string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		go producter(&count)
	}
	for i := 0; i < 5; i++ {
		go consumer(&count)
	}
	<-quit //这个管道是故意阻塞在这里，防止主进程结束
}
