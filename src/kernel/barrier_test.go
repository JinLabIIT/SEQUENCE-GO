package kernel

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func random_wait1(br *Barrier, wg *sync.WaitGroup, wants []int64, i int, nextStop uint64, min *uint64) {
	n := rand.Intn(1) //changed
	fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	*min, _ = br.waitEventExchange(nextStop, 10)
	wants[i] = time.Now().Unix()
	wg.Done()
}

func random_wait2(br *Barrier, wg *sync.WaitGroup, wants []int64, i int) {
	n := rand.Intn(1) //changed
	fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	br.waitExecution()
	wants[i] = time.Now().Unix()
	wg.Done()
}

func TestBarrier_waitEventExchange(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"test1", 2},
		{"test2", 4},
		{"test3", 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Barrier{n: tt.n}
			b.Init()
			n := tt.n
			var min uint64
			nextStop := make([]uint64, tt.n)
			test_min := uint64(math.MaxUint64)
			for i := 0; i < tt.n; i++ {
				random := rand.Uint64()
				nextStop[i] = random
				if test_min > random {
					test_min = random
				}
			}
			var wg sync.WaitGroup
			wants := make([]int64, n)
			for i := 0; i < n; i++ {
				wg.Add(1)
				go random_wait1(&b, &wg, wants, i, nextStop[i], &min)
			}
			wg.Wait()
			if n != b.c {
				t.Error("Counter error in barrier, expected: ", n, " get: ", b.c)
			}
			if min != test_min {
				fmt.Println(nextStop)
				fmt.Println(min)
				fmt.Println(test_min)
				t.Error("minimum does not match")
			}
			for i := 1; i < n; i++ {
				if wants[i] != wants[0] {
					t.Error("different finish time: ", wants)
				}
			}
		})
	}
}

func TestBarrier_waitExecution(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"test1", 2},
		{"test2", 4},
		{"test3", 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Barrier{n: tt.n}
			b.Init()
			n := tt.n
			b.c = n
			var wg sync.WaitGroup
			wants := make([]int64, n)
			for i := 0; i < n; i++ {
				wg.Add(1)
				go random_wait2(&b, &wg, wants, i)
			}
			wg.Wait()
			if b.c != 0 {
				t.Error("Counter error in barrier, expected: 0 get: ", b.c)
			}

			for i := 1; i < n; i++ {
				if wants[i] != wants[0] {
					t.Error("different finish time: ", wants)
				}
			}
		})
	}
}
