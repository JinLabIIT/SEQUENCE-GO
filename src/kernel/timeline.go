package kernel

import (
	"fmt"
	"math"
	"os"
	"sync"
)

type Timeline struct {
	Name           string // timeline name
	time           uint64
	events         EventList
	endTime        uint64 // execution time: [0, endTime)
	nextStopTime   uint64
	eventBuffer    EventBuffer
	otherTimeline  []*Timeline
	LookAhead      uint64
	executedEvent  uint64
	scheduledEvent uint64
}

func (t *Timeline) Init(lookahead, endTime uint64) {
	t.eventBuffer = make(EventBuffer)
	t.events = EventList{}
	t.executedEvent = 0
	t.scheduledEvent = 0
	t.LookAhead = lookahead
	t.endTime = endTime
}

func (t *Timeline) SetEndTime(endTime uint64) {
	t.endTime = endTime
}

func (t *Timeline) Now() uint64 {
	return t.time
}

func (t *Timeline) Schedule(event *Event) {
	if t.time > event.Time {
		fmt.Println("ERROR: cannot schedule an event in the past time")
		os.Exit(3) //cannot schedule an event in the past time
	}
	if t == event.Process.Owner {
		t.scheduledEvent += 1
		t.events.push(event)
	} else {
		t.eventBuffer.push(event)
	}
}

// get events in the event buffer
func (t *Timeline) getCrossTimelineEvents() {
	for _, timeline := range t.otherTimeline {
		if timeline.eventBuffer[t] == nil || timeline.eventBuffer[t].size() == 0 {
			continue
		}
		t.scheduledEvent += uint64(timeline.eventBuffer[t].size())
		t.events.merge(*timeline.eventBuffer[t])
	}
}

func (t *Timeline) minNextStopTime() uint64 {
	if t.events.size() == 0 { //Eventlist is empty in this timeline
		return uint64(math.MaxInt64)
	}
	return t.events.top().Time + t.LookAhead
}

func (t *Timeline) updateNextStopTime(nextStop uint64) {
	t.nextStopTime = nextStop
	if t.nextStopTime > t.endTime || len(t.otherTimeline) == 1 {
		t.nextStopTime = t.endTime
	}
}

func (t *Timeline) syncWindow() {
	for t.events.size() != 0 && t.events.top().Time < t.nextStopTime {
		if t.events.size() == 0 {
			break
		}
		event := t.events.pop()
		t.time = event.Time
		t.executedEvent += 1
		event.Process.run()
	}
}

func (t *Timeline) cleanEvenbuffer() {
	t.eventBuffer = make(EventBuffer)
}

func (t *Timeline) run(br *Barrier, wg *sync.WaitGroup) {
	for {
		var maxListSize int
		t.getCrossTimelineEvents()
		nextStop := t.minNextStopTime()
		nextStop, maxListSize = br.waitEventExchange(nextStop, t.events.size())
		if maxListSize == 0 {
			break
		}
		t.updateNextStopTime(nextStop)
		t.cleanEvenbuffer()
		t.syncWindow()
		if t.nextStopTime == t.endTime {
			break
		}
		br.waitExecution()
	}
	wg.Done()
}
