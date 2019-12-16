package kernel

import (
	"sync"
)

//seed

func Run(timelineList []*Timeline, totalNodes int) {
	br := Barrier{}
	br.n = totalNodes
	br.Init()
	for _, timeline := range timelineList {
		eventbuffer := make(EventBuffer)
		timeline.otherTimeline = timelineList
		timeline.eventBuffer = eventbuffer
		timeline.events = EventList{}
	}
	var wg sync.WaitGroup
	for _, timeline := range timelineList {
		wg.Add(1)
		go timeline.run(&br, &wg)
	}
	wg.Wait()
}
