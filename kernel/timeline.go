package kernel

type Timeline struct {
	time         uint64
	events       EventList
	entities     []Entity
	endTime      uint64
	nextStopTime uint64
	eventbuffer  EventBuffer
	otherTimeline []*Timeline
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
	for _, timeline := range t.otherTimeline{
		t.events.merge(*t.eventbuffer[timeline])
		t.eventbuffer.clean()
	}
}

func (t *Timeline) updateNextStopTime() {
	// TODO
}

func (t *Timeline) syncWindow() {
	// method sync_window() is called to do all the processing
	// associated with the window.
	// TODO
}

func (t *Timeline) run() {
	for {
		// TODO
		t.getCrossTimelineEvents()
		t.updateNextStopTime()
		t.syncWindow()
	}
}
