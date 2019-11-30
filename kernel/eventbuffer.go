package kernel

type EventBuffer map[*Timeline]*EventList

func (eb *EventBuffer) push(e *Event) {
	// TODO Bo
}

func (eb *EventBuffer) clean() {
	// TODO Bo
}
