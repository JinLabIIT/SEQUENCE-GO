package kernel

type EventBuffer map[*Timeline][]*Event

func (eb *EventBuffer) push(e *Event) {
	own := e.Process.Owner
	if (*eb)[own] == nil {
		tmp_el := make([]*Event, 0, 0)
		// tmp_el := EventList{make([]*Event, 0, 100000)}
		(*eb)[own] = tmp_el
	}
	(*eb)[own] = append((*eb)[own], e)
}

func (eb *EventBuffer) clean() {
	//(*eb)[timeline] = &EventList{}
	// timeline.eventBuffer = make(EventBuffer)
	for own, _ := range *eb {
		(*eb)[own] = nil
	}
}

func (eb *EventBuffer) size() int {
	var size int
	for _, evenlist := range *eb {
		if evenlist != nil {
			size += len(evenlist)
		}
	}
	return size
}
