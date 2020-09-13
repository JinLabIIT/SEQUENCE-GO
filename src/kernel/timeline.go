package kernel

import (
	"bufio"
	"fmt"
	"math"
	"os"
	_ "reflect"
	_ "runtime"
	"strconv"
	"sync"
	"time"
	_ "time"
)

type Timeline struct {
	Name           string // timeline name
	dir            string
	time           uint64
	events         EventList
	endTime        uint64 // execution time: [0, endTime)
	nextStopTime   uint64
	eventBuffer    EventBuffer
	otherTimeline  []*Timeline
	LookAhead      uint64
	executedEvent  uint64
	scheduledEvent uint64
	SyncCounter    uint64
	EventPool      *sync.Pool
	PhotonPool     *sync.Pool
	lucky          int
	luckyCounter   int
	past           int64
}

func (t *Timeline) writeToFile2(f *os.File) {
	now := time.Now().UnixNano()
	datawrite := bufio.NewWriter(f)
	_, _ = datawrite.WriteString(strconv.FormatInt(now-t.past, 10) + "," + strconv.Itoa(t.luckyCounter))
	datawrite.WriteString("\n")
	datawrite.Flush()
}

func (t *Timeline) Init(lookahead, endTime uint64) {
	t.eventBuffer = make(EventBuffer)
	t.events = EventList{make([]*Event, 0, 0)}
	t.executedEvent = 0
	t.scheduledEvent = 0
	t.LookAhead = lookahead
	t.endTime = endTime
}

func (t *Timeline) SetEndTime(endTime uint64) {
	t.endTime = endTime
}

func (t *Timeline) Now() uint64 {
	return t.time
}

func (t *Timeline) Schedule(event *Event) {
	if t.time > event.Time {
		panic("ERROR: cannot schedule an event in the past time")
	}
	if t == event.Process.Owner {
		t.scheduledEvent += 1
		t.events.push(event)
	} else {
		t.eventBuffer.push(event)
	}
}

func (t *Timeline) getCrossTimelineEvents() {
	var eb []*EventList
	eb = append(eb, &t.events)

	for i := 0; i < len(t.otherTimeline); i++ {
		var tmp *EventList
		if t.otherTimeline[i].eventBuffer[t] == nil || t.otherTimeline[i].eventBuffer[t].size() == 0 {
			continue
		}
		tmp = t.otherTimeline[i].eventBuffer[t]
		t.scheduledEvent += uint64(tmp.size())
		eb = append(eb, tmp)
	}
	for len(eb) != 1 {
		//memory question
		var eb2 []*EventList
		for i := 0; i < len(eb); i += 2 {
			if i+1 < len(eb) {
				eb[i].merge(*(eb[i+1]))
				eb2 = append(eb2, eb[i])
			} else {
				eb2 = append(eb2, eb[i])
			}
		}
		eb = eb2
	}
}

func (t *Timeline) minNextStopTime() uint64 {
	if t.events.size() == 0 { //Eventlist is empty in this timeline
		return uint64(math.MaxInt64)
	}
	return t.events.top().Time + t.LookAhead
}

func (t *Timeline) updateNextStopTime(nextStop uint64) {
	t.nextStopTime = nextStop
	if t.nextStopTime > t.endTime || len(t.otherTimeline) == 1 {
		t.nextStopTime = t.endTime
	}
}

func (t *Timeline) syncWindow() {
	//past:= time.Now().UnixNano()
	//NoEvents := t.executedEvent
	for t.events.size() != 0 && t.events.top().Time < t.nextStopTime {
		event := t.events.pop()
		if event.Time < t.time {
			err_msg := fmt.Sprint("running an earlier event now: ", t.time, " event: ", event.Time)
			panic(err_msg)
		}
		t.time = event.Time
		t.executedEvent += 1
		t.luckyCounter += 1
		event.Process.run()
		if event.Process.Message["StateList"] != nil{
			event.Process.Message["StateList"] = nil
		}
		t.EventPool.Put(event)
	}
}

func (t *Timeline) cleanEvenbuffer() {
	t.eventBuffer.clean()
}

func (t *Timeline) run(br *Barrier, wg *sync.WaitGroup) {
	/*filename := "thread2/data"+ t.Name +".txt"
	os.Create(filename)
	f,err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE,0644)
	if err != nil{
		fmt.Println(err)
		f.Close()
	}*/
	for {
		t.SyncCounter += 1
		var maxListSize int
		t.getCrossTimelineEvents()
		nextStop := t.minNextStopTime()
		nextStop, maxListSize = br.waitEventExchange(nextStop, t.events.size())
		if maxListSize == 0 {
			break
		}
		t.updateNextStopTime(nextStop)
		t.cleanEvenbuffer()
		t.syncWindow()
		if t.nextStopTime == t.endTime {
			break
		}
		br.waitExecution()
	}
	//f.Close()
	wg.Done()
}
