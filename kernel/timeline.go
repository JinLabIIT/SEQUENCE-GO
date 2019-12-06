package kernel

type Timeline struct {
	time          uint64
	events        EventList
	entities      []Entity
	endTime       uint64
	nextStopTime  uint64
	eventbuffer   EventBuffer
	otherTimeline []*Timeline
	look_head     uint64
}

func (t *Timeline) init() {
	for _, entity := range t.entities {
		entity.init()
	}
}

func (t *Timeline) setStopTime(stop_time uint64) {
	t.endTime = stop_time
}

func (t *Timeline) setEntities(entities []Entity) {
	t.entities = entities
}

func (t *Timeline) Now() uint64 {
	return t.time
}

func (t *Timeline) Schedule(event *Event) {
	if t == event.process.owner.timeline {
		t.events.push(event)
	} else {
		t.eventbuffer.push(event)
	}
}

// get events in the event buffer
func (t *Timeline) getCrossTimelineEvents() {
	tmp := 0
	index := 0
	for _, timeline := range t.otherTimeline {
		index++
		if timeline.eventbuffer[t] == nil {
			continue
		}
		tmp = tmp + timeline.eventbuffer[t].size()
		t.events.merge(*timeline.eventbuffer[t])
	}
	//fmt.Println("here is getcrosstimelineevents and tmp is: ")
}

func (t *Timeline) minNextStopTime() uint64 {
	return t.events.top().time + t.look_head
}

func (t *Timeline) updateNextStopTime(nextStop uint64) {
	t.nextStopTime = nextStop

}

func (t *Timeline) syncWindow() {
	for t.time < t.nextStopTime {
		event := t.events.top()
		if event.time > t.nextStopTime {
			return
		}
		t.time = event.time
		event = t.events.pop()
		event.process.run()
	}
}

func (t *Timeline) run(br *Barrier) {
	for {
		t.getCrossTimelineEvents()
		nextStop := t.minNextStopTime()
		nextStop = br.waitEventExchange(nextStop)
		t.updateNextStopTime(nextStop)
		t.eventbuffer.clean(t)
		t.syncWindow()
		br.waitExecution()
	}
}
