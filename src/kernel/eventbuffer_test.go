package kernel

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestEvenbuffer_push(t *testing.T) {
	eventbuffer := make(EventBuffer)
	n := 30                    //No. timeline
	tl := make([]*Timeline, n) //init timeline
	for i := 0; i < n; i++ {
		tmp_tl := Timeline{time: uint64(i)}
		tl[i] = &tmp_tl
	}

	//init eventbuffer
	eventlist := EventList{}
	a := n * 100 //No. event
	for i := 0; i < a; i++ {
		random := rand.Intn(n)
		entity := Entity{timeline: tl[random]}
		process := Process{owner: &entity}
		event := &Event{time: uint64(rand.Intn(10)), priority: rand.Intn(10), process: &process}
		eventlist.push(event)
		eventbuffer.push(event)
	}

	for _, timeline := range tl {
		for eventbuffer[timeline].size() > 0 {
			t.Run("push test", func(t *testing.T) {
				event := eventbuffer[timeline].pop()
				if !reflect.DeepEqual(event.process.owner.timeline, timeline) {
					t.Errorf("something wrong")
				}
			})
		}

	}
}
