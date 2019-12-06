package kernel

type EventBuffer map[*Timeline]*EventList

func (eb *EventBuffer) push(e *Event) {
	own := e.process.owner.timeline
	if (*eb)[own] == nil {
		tmp_el := EventList{}
		(*eb)[own] = &tmp_el
	}
	evenlist := (*eb)[own]
	evenlist.push(e)
}

func (eb *EventBuffer) cleant() {
	for _, evenlist := range *eb {
		*evenlist = EventList{}
	}
}

func (eb *EventBuffer) clean(timeline *Timeline) {
	(*eb)[timeline] = &EventList{}
}
