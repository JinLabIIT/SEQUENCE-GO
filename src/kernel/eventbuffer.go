package kernel

type EventBuffer map[*Timeline]*EventList

func (eb *EventBuffer) push(e *Event) {
	own := e.Process.Owner
	if (*eb)[own] == nil {
		tmp_el := EventList{make([]*Event, 0, 0)}
		// tmp_el := EventList{make([]*Event, 0, 100000)}
		(*eb)[own] = &tmp_el
	}
	evenlist := (*eb)[own]
	evenlist.push(e)
}

func (eb *EventBuffer) clean() {
	//(*eb)[timeline] = &EventList{}
	// timeline.eventBuffer = make(EventBuffer)
	for _, eventlist := range *eb {
		eventlist.events = eventlist.events[:0]
	}
}

func (eb *EventBuffer) size() int {
	var size int
	for _, evenlist := range *eb {
		size += evenlist.size()
	}
	return size
}
