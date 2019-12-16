package kernel

type EventBuffer map[*Timeline]*EventList

func (eb *EventBuffer) push(e *Event) {
	own := e.Process.Owner.Timeline
	if (*eb)[own] == nil {
		tmp_el := EventList{}
		(*eb)[own] = &tmp_el
	}
	evenlist := (*eb)[own]
	evenlist.push(e)
}

func (eb *EventBuffer) clean(timeline *Timeline) {
	//(*eb)[timeline] = &EventList{}
	timeline.eventBuffer = make(EventBuffer)
}

func (eb *EventBuffer) size() int {
	var size int
	for _, evenlist := range *eb {
		size += evenlist.size()
	}
	return size
}
