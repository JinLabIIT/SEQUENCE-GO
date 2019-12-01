// github.com/mlowicki/barrier2
package kernel

import "sync"

type Barrier struct {
	n      int
	c      int
	m      sync.Mutex
	before chan int
	after  chan int
}

func (b *Barrier) Init() {
	b.before = make(chan int, b.n)
	b.after = make(chan int, b.n)
}

func (b *Barrier) waitEventExchange() {
	b.m.Lock()
	b.c += 1
	if b.c == b.n {
		// open 2nd gate
		for i := 0; i < b.n; i++ {
			b.before <- 1
		}
	}
	b.m.Unlock()
	<-b.before
}
func (b *Barrier) waitExecution() {
	b.m.Lock()
	b.c -= 1
	if b.c == 0 {
		// open 1st gate
		for i := 0; i < b.n; i++ {
			b.after <- 1
		}
	}
	b.m.Unlock()
	<-b.after
}
