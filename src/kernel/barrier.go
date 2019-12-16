// github.com/mlowicki/barrier2
package kernel

import (
	"math"
	"sync"
)

type Barrier struct {
	n         int
	c         int
	m         sync.Mutex
	before    chan int
	after     chan int
	next_stop uint64
	maxSize   int
}

func (b *Barrier) Init() {
	b.before = make(chan int, b.n)
	b.after = make(chan int, b.n)
	b.next_stop = uint64(math.MaxInt64)
}

func (b *Barrier) waitEventExchange(_next_stop uint64, size int) (uint64, int) {
	b.m.Lock()
	b.c += 1
	b.next_stop = min(_next_stop, b.next_stop)
	b.maxSize = max(b.maxSize, size)
	if b.c == b.n {
		for i := 0; i < b.n; i++ {
			b.before <- 1
		}
	}
	b.m.Unlock()
	<-b.before
	return b.next_stop, b.maxSize
}

func (b *Barrier) waitExecution() {
	b.m.Lock()
	b.c -= 1
	if b.c == 0 {
		b.next_stop = uint64(math.MaxInt64)
		b.maxSize = 0
		// open 1st gate
		for i := 0; i < b.n; i++ {
			b.after <- 1
		}
	}
	b.m.Unlock()
	<-b.after
}

func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
