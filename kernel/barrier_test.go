package kernel

import (
	"sync"
	"testing"
)

func TestBarrier_waitEventExchange(t *testing.T) {
	type fields struct {
		c      int
		n      int
		m      sync.Mutex
		before chan int
		after  chan int
	}

	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Barrier{
				c:      tt.fields.c,
				n:      tt.fields.n,
				m:      tt.fields.m,
				before: tt.fields.before,
				after:  tt.fields.after,
			}
		})
	}
}

func TestBarrier_waitExecution(t *testing.T) {
	type fields struct {
		c      int
		n      int
		m      sync.Mutex
		before chan int
		after  chan int
	}
	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Barrier{
				c:      tt.fields.c,
				n:      tt.fields.n,
				m:      tt.fields.m,
				before: tt.fields.before,
				after:  tt.fields.after,
			}
		})
	}
}
