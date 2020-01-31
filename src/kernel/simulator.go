package kernel

import (
	"fmt"
	"sync"
)

func Run(timelineList []*Timeline) {
	br := Barrier{}
	br.n = len(timelineList)
	br.Init()
	for _, timeline := range timelineList {
		timeline.otherTimeline = timelineList
		//timeline.events = EventList{make([]*Event,0,2000)}
	}
	var wg sync.WaitGroup
	for _, timeline := range timelineList {
		wg.Add(1)
		go timeline.run(&br, &wg)
	}
	wg.Wait()
	timelineReport(timelineList)
}

func timelineReport(timelineList []*Timeline) {
	for _, timeline := range timelineList {
		fmt.Println(timeline.Name, "number of executed event:", timeline.executedEvent, "number of scheduled event:", timeline.scheduledEvent)
	}
}
