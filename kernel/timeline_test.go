package kernel

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func initTimeline(n int) []*Timeline {
	tl := make([]*Timeline, n)
	for i := 0; i < n; i++ {
		tl[i] = &Timeline{time: uint64(i)}
		tl[i].otherTimeline = tl
	}
	return tl
}

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
	n := 20               //No. timeline
	a := 1000             //No. event
	tl := initTimeline(n) //init timeline
	eventsize := make([]int, n)
	eventbufferlist := make([]EventBuffer, n)
	rd := make([]int, n*a)
	for i := 0; i < n; i++ {
		eventbuffer := make(EventBuffer)
		tl[i].otherTimeline = tl
		for j := 0; j < a; j++ {
			random := rand.Intn(n)
			rd[(i+1)*j] = random
			entity := Entity{timeline: tl[random]}
			eventsize[random] += 1
			process := Process{owner: &entity}
			event := &Event{time: uint64(rand.Intn(10)), priority: rand.Intn(10), process: &process}
			eventbuffer.push(event)
		}
		eventbufferlist[i] = eventbuffer
		tl[i].eventbuffer = eventbuffer
		fmt.Println(rd)
	}

	size := make([]int, n)
	for i := 0; i < n; i++ {
		tl[i].getCrossTimelineEvents()
		size[i] = tl[i].events.size()
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
func TestTimeline_syncWindow(t *testing.T) {

}
