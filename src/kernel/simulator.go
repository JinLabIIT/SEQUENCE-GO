package kernel

import (
	"sync"
)

//seed

func Run(timelineList []*Timeline) {
	br := Barrier{}
	br.n = len(timelineList)
	br.Init()
	for _, timeline := range timelineList {
		timeline.otherTimeline = timelineList
		timeline.init()
	}
	var wg sync.WaitGroup
	for _, timeline := range timelineList {
		wg.Add(1)
		go timeline.run(&br, &wg)
	}
	wg.Wait()
}
