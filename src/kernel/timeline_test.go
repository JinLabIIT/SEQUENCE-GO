package kernel

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

type Dummmy struct {
	name string
}

func (dummy *Dummmy) dummyFunction(message Message) {
	//do nothing
	fmt.Println("wahaha")
}

func initTimeline(n int, nextStop []uint64) []*Timeline {
	tl := make([]*Timeline, n)
	for i := 0; i < n; i++ {
		tl[i] = &Timeline{time: uint64(i)}
		tl[i].otherTimeline = tl
		tl[i].nextStopTime = nextStop[i] + 2
	}
	return tl
}

func createEvent(timeline *Timeline, time uint64, priority uint) *Event {
	entity := Entity{Timeline: timeline}
	process := Process{Owner: &entity}
	event := &Event{Time: time, Priority: priority, Process: &process}
	return event
}

func TestTimeline_Schedule(t *testing.T) {
	fmt.Println("TestTimeline_Schedule starts")
	eventlist := EventList{}
	timeline := Timeline{time: 0, events: eventlist}
	test_eventlist := EventList{}
	n := 1000
	for i := 0; i < n; i++ {
		event := createEvent(&timeline, uint64(rand.Intn(10)+1), uint(rand.Intn(10)))
		timeline.Schedule(event)
		test_eventlist.push(event)
	}
	for test_eventlist.size() > 0 {
		t.Run("TestTimeline_Schedule", func(t *testing.T) {
			event_test := test_eventlist.pop()
			event := timeline.events.pop()
			if !reflect.DeepEqual(event_test, event) {
				t.Errorf("something wrong")
			}
		})
	}
}

func TestTimeline_getCrossTimelineEvents(t *testing.T) {
	n := 2 //No. timeline
	a := 1 //No. event
	nextStop := make([]uint64, n)
	tl := initTimeline(n, nextStop) //init timeline
	rd := make([]int, n*a)
	for i := 0; i < n; i++ {
		eventbuffer := make(EventBuffer)
		tl[i].otherTimeline = tl
		for j := 0; j < a; j++ {
			random := rand.Intn(n)
			rd[(i+1)*j] = random
			event := createEvent(tl[random], uint64(rand.Intn(10)), uint(rand.Intn(10)))
			eventbuffer.push(event)
		}
		tl[i].eventBuffer = eventbuffer
		//fmt.Println(rd)
	}

	//size := make([]int, n)
	for i := 0; i < n; i++ {
		tl[i].getCrossTimelineEvents()
		//size[i] = tl[i].events.size()
	}

	for _, timeline := range tl {
		for timeline.events.size() > 0 {
			t.Run("getCrossTimelineEvents", func(t *testing.T) {
				event := timeline.events.pop()
				if !reflect.DeepEqual(event.Process.Owner.Timeline, timeline) {
					t.Errorf("something wrong")
				}
			})
		}
	}
}
func TestTimeline_syncWindow(t *testing.T) {
	n := 10
	a := 100
	nextStop := make([]uint64, n)
	for i := 0; i < n; i++ {
		randomTime := rand.Intn(30)
		nextStop[i] = uint64(randomTime) + 10
	}
	tl := initTimeline(n, nextStop)
	for i := 0; i < n; i++ {
		timeline := tl[i]
		for j := 0; j < a; j++ {
			d1 := Dummmy{"alice"}
			event := createEvent(timeline, uint64(rand.Intn(40))+timeline.time, uint(rand.Intn(40)))
			event.Process.Fnptr = d1.dummyFunction
			event.Process.Message = Message{"info": "???"}
			timeline.Schedule(event)
		}
		timeline.Schedule(createEvent(timeline, 41, 0))
	}
	for i := 0; i < n; i++ {
		t.Run("Timeline_syncWindow", func(t *testing.T) {
			tl[i].syncWindow()
			if tl[i].time > tl[i].nextStopTime {
				t.Errorf("something wrong")
				fmt.Print("tl.time: ")
				fmt.Println(tl[i].time)
				fmt.Print("tl.nextStop time: ")
				fmt.Println(tl[i].nextStopTime)
			}
		})
	}
}
