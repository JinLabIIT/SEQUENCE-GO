package kernel

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestTimeline_Schedule(t *testing.T) {
	fmt.Println("TestTimeline_Schedule starts")
	eventlist := EventList{}
	timeline := Timeline{time: 0, events: eventlist}
	test_eventlist := EventList{}
	n := 1000
	for n > 0 {
		entity := Entity{timeline: &timeline}
		process := Process{owner: &entity}
		event := &Event{time: uint64(rand.Intn(10)), priority: rand.Intn(10), process: &process}
		timeline.Schedule(event)
		test_eventlist.push(event)
		n--
	}
}
func TestTimeline_getCrossTimelineEvents(t *testing.T) {
	n := 30                    //No. timeline
	a := 100                   //No. event
	tl := make([]*Timeline, n) //init timeline

	for i := 0; i < n; i++ {
		eventbuffer := make(EventBuffer)
		tmp_tl := Timeline{time: uint64(i)}
		tl[i] = &tmp_tl
		tl[i].otherTimeline = tl
		for j := 0; j < a; j++ {
			random := rand.Intn(n)
			entity := Entity{timeline: tl[random]}
			process := Process{owner: &entity}
			event := &Event{time: uint64(rand.Intn(10)), priority: rand.Intn(10), process: &process}
			eventbuffer.push(event)
		}
		tl[i].eventbuffer = eventbuffer
	}

	for i := 0; i < n; i++ {
		tl[i].getCrossTimelineEvents()
	}

	for _, timeline := range tl {
		for timeline.events.size() > 0 {
			t.Run("getCrossTimelineEvents", func(t *testing.T) {
				event := timeline.events.pop()
				if !reflect.DeepEqual(event.process.owner.timeline, timeline) {
					t.Errorf("something wrong")
				}
			})
		}
	}
}
