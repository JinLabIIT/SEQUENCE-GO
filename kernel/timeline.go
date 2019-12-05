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
	for _, timeline := range t.otherTimeline {
		t.events.merge(*t.eventbuffer[timeline])
	}
}

func (t *Timeline) uploadNextStopTime(tmp map[*Timeline]uint64) {
	next_stop := t.events.top().time + t.look_head
	tmp[t] = next_stop
}

func (t *Timeline) updateNextStopTime() {
	//next_stop := t.events.top().time+t.look_head

}

func (t *Timeline) syncWindow() {
	for t.time < t.endTime {
		event := t.events.top()
		if event.time > t.endTime {
			return
		}
		t.time = event.time
		event = t.events.pop()
		event.process.run()
	}
}

func (t *Timeline) run(br *Barrier, tmp map[*Timeline]uint64) {
	for {
		// TODO
		t.getCrossTimelineEvents()
		t.uploadNextStopTime(tmp)
		//nxt = br.waitEventExchange(6)
		t.updateNextStopTime()
		t.syncWindow()
		br.waitExecution()
	}
}
