package kernel

import (
	"fmt"
	"math"
	_ "reflect"
	_ "runtime"
	"sync"
	_ "time"
)

type Timeline struct {
	Name           string // timeline name
	time           uint64
	events         *LadderQ
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
	t.events = NewLadderQ(50, 8)
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
		panic("ERROR: cannot schedule an event in the past time")
	}
	if t == event.Process.Owner {
		t.scheduledEvent += 1
		t.events.Push(event)
	} else {
		t.eventBuffer.push(event)
	}
}

func (t *Timeline) getCrossTimelineEvents() {
	for i := 0; i < len(t.otherTimeline); i++ {
		if t.otherTimeline[i].eventBuffer[t] == nil {
			continue
		}
		for j := 0; j < len(t.otherTimeline[i].eventBuffer[t]); j++ {
			t.events.Push(t.otherTimeline[i].eventBuffer[t][j])
		}
	}
}

func (t *Timeline) minNextStopTime() uint64 {
	if t.events.Size() == 0 { //Eventlist is empty in this timeline
		return uint64(math.MaxInt64)
	}
	return t.events.Top().Time + t.LookAhead
}

func (t *Timeline) updateNextStopTime(nextStop uint64) {
	t.nextStopTime = nextStop
	if t.nextStopTime > t.endTime || len(t.otherTimeline) == 1 {
		t.nextStopTime = t.endTime
	}
}

func (t *Timeline) syncWindow() {
	//past:= time.Now().UnixNano()
	//NoEvents := t.executedEvent
	for t.events.Size() != 0 && t.events.Top().Time < t.nextStopTime {
		event := t.events.Pop()
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
		nextStop, maxListSize = br.waitEventExchange(nextStop, t.events.Size())
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
