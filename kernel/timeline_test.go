package kernel

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestTimeline_Schedule(t *testing.T)  {
	fmt.Println("TestTimeline_Schedule starts")
	eventlist := EventList{}
	timeline := Timeline{time:0,events:eventlist}
	test_eventlist := EventList{}
	n := 1000
	for ;n>0;{
		entity := Entity{timeline: &timeline}
		process := Process{owner: &entity}
		event := &Event{time:uint64(rand.Intn(10)),priority:rand.Intn(10),process:&process}
		timeline.Schedule(event)
		test_eventlist.push(event)
		n--
	}
}
func TestTimeline_getCrossTimelineEvents(t *testing.T) {
	eventbuffer := make(EventBuffer)

	n := 30 //No. timeline
	tl := make([]*Timeline,n)//init timeline
	for i:=0;i<n;i++{
		tmp_tl :=Timeline{time: uint64(i)}
		tl[i] = &tmp_tl
	}

	//init eventbuffer
	a := n*100 //No. event
	for i:=0;i<a;i++{
		random := rand.Intn(n)
		entity := Entity{timeline: tl[random]}
		process := Process{owner: &entity}
		event := &Event{time:uint64(rand.Intn(10)),priority:rand.Intn(10),process:&process}
		eventbuffer.push(event)
	}

}