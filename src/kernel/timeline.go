package kernel

import (
	"fmt"
	"os"
)

type Timeline struct {
	Name          string // timeline name
	time          uint64
	events        EventList
	entities      []Entity
	endTime       uint64
	nextStopTime  uint64
	eventBuffer   EventBuffer
	otherTimeline []*Timeline
	LookAhead     uint64
}

func (t *Timeline) init() {
	for _, entity := range t.entities {
		entity.init()
	}
}

func (t *Timeline) SetEndTime(endTime uint64) {
	t.endTime = endTime
}

func (t *Timeline) setEntities(entities []Entity) {
	t.entities = entities
}

func (t *Timeline) Now() uint64 {
	return t.time
}

func (t *Timeline) Schedule(event *Event) {
	if t.time > event.Time {
		fmt.Println("ERROR: cannot schedule an event in the past time")
		os.Exit(3) //cannot schedule an event in the past time
	}
	if t == event.Process.Owner.Timeline {
		t.events.push(event)
	} else {
		t.eventBuffer.push(event)
	}
}

// get events in the event buffer
func (t *Timeline) getCrossTimelineEvents() {
	for _, timeline := range t.otherTimeline {
		if timeline.eventBuffer[t] == nil {
			continue
		}
		t.events.merge(*timeline.eventBuffer[t])
	}
}

func (t *Timeline) minNextStopTime() uint64 {
	return t.events.top().Time + t.LookAhead
}

func (t *Timeline) updateNextStopTime(nextStop uint64) {
	t.nextStopTime = nextStop

}

func (t *Timeline) syncWindow() {

	for t.events.top().Time < t.nextStopTime {
		event := t.events.pop()
		t.time = event.Time
		event.Process.run()
	}
}

func (t *Timeline) run(br *Barrier) {
	for {
		var flag int
		t.getCrossTimelineEvents()
		nextStop := t.minNextStopTime()
		nextStop, flag = br.waitEventExchange(nextStop, t.events.size())
		if flag == -1 {
			break
		}
		t.updateNextStopTime(nextStop)
		t.eventBuffer.clean(t)
		t.syncWindow()
		br.waitExecution()
	}
}
