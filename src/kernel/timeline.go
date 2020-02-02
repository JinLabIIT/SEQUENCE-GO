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
	SyncCounter    uint64
}

func (t *Timeline) Init(lookahead, endTime uint64) {
	t.eventBuffer = make(EventBuffer)
	t.events = EventList{make([]*Event, 0, 0)}
	// t.events = EventList{make([]*Event, 0, 100000)}
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

func (t *Timeline) getCrossTimelineEvents() {
	//fmt.Println("before", t.Name, t.events.size())
	var eb []*EventList
	eb = append(eb, &t.events)

	for i := 0; i < len(t.otherTimeline); i++ {
		var tmp *EventList
		if t.otherTimeline[i].eventBuffer[t] == nil || t.otherTimeline[i].eventBuffer[t].size() == 0 {
			continue
		}
		tmp = t.otherTimeline[i].eventBuffer[t]
		t.scheduledEvent += uint64(tmp.size())
		eb = append(eb, tmp)
	}
	//fmt.Println("    ", t.Name, eb[len(t.otherTimeline)-1].size())
	for len(eb) != 1 {
		//memory question
		var eb2 []*EventList
		for i := 0; i < len(eb); i += 2 {
			if i+1 < len(eb) {
				eb[i].merge(*(eb[i+1]))
				eb2 = append(eb2, eb[i])
			} else {
				eb2 = append(eb2, eb[i])
			}
		}
		eb = eb2
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
		event := t.events.pop()
		if event.Time < t.time {
			err_msg := fmt.Sprint("running an earlier event now: ", t.time, " event: ", event.Time)
			panic(err_msg)
		}
		t.time = event.Time
		t.executedEvent += 1
		event.Process.run()
	}
}

func (t *Timeline) cleanEvenbuffer() {
	t.eventBuffer.clean()
}

func (t *Timeline) run(br *Barrier, wg *sync.WaitGroup) {
	for {
		t.SyncCounter += 1
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
