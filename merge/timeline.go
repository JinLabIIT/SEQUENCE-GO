package kernel

type Timeline struct {
	time          uint64
	events        EventList
	entities      []Entity
	endTime       uint64
	nextStopTime  uint64
	activeChannel *InChannel
}

func (t *Timeline) init() {
	t.endTime = 0
	t.time = 0
	t.entityInit()
}

func (t *Timeline) setStopTime(stop_time uint64) {
	t.endTime = stop_time
}

func (t *Timeline) entityInit() {
	for _, entity := range t.entities {
		entity.init()
	}
}

func (t *Timeline) setEntities(entities []Entity) {
	t.entities = entities
}

func (t *Timeline) now() uint64 {
	return t.time
}

func (t *Timeline) schedule(time uint64, process *Process) {
	event := &Event{time: time, process: process}
	t.events.push(event)
}

func (t *Timeline) getCrossTimelineEvents() {
	// TODO
}

func (t *Timeline) updateNextStopTime() {
	// TODO
}

func (t *Timeline) sync_window() {
	// method sync_window() is called to do all the processing
	// associated with the window.
	// TODO
}

func (t *Timeline) run() {
	for {
		// TODO
		t.getCrossTimelineEvents()
		t.updateNextStopTime()
		t.sync_window()
	}
}
