package kernel

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestEvenbuffer_push(t *testing.T) {
	eventbuffer := make(EventBuffer)
	n := 3                     //No. timeline
	tl := make([]*Timeline, n) //init timeline
	for i := 0; i < n; i++ {
		tmp_tl := Timeline{time: uint64(i)}
		tl[i] = &tmp_tl
	}

	//init eventbuffer
	eventlist := EventList{}
	a := n * 10 //No. event
	for i := 0; i < a; i++ {
		random := rand.Intn(n)
		process := Process{Owner: tl[random]}
		event := &Event{Time: uint64(rand.Intn(10)), Priority: uint(rand.Intn(10)), Process: &process}
		eventlist.push(event)
		eventbuffer.push(event)
	}

	for _, timeline := range tl {
		for eventbuffer[timeline].size() > 0 {
			t.Run("push test", func(t *testing.T) {
				event := eventbuffer[timeline].pop()
				if !reflect.DeepEqual(event.Process.Owner, timeline) {
					t.Errorf("something wrong")
				}
			})
		}

	}
}
