package kernel

import (
	"fmt"
	"math"
	"math/rand"
	_ "reflect"
	_ "runtime"
	"sync"
	"time"
	_ "time"
)

type Timeline struct {
	Name           string // timeline name
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
	Type_counter   map[string]int
	Type_timer     map[string]float32
	Type_sample    map[string][]float32
}

func (t *Timeline) Init(lookahead, endTime uint64) {
	t.eventBuffer = make(EventBuffer)
	t.events = EventList{make([]*Event, 0, 0)}
	t.executedEvent = 0
	t.scheduledEvent = 0
	t.LookAhead = lookahead
	t.endTime = endTime
	t.Type_counter = map[string]int{"qc": 0, "cc": 0, "other": 0}
	t.Type_timer = map[string]float32{"qc": 0, "cc": 0, "other": 0}
	t.Type_sample = map[string][]float32{"qc": []float32{}, "cc": []float32{}, "other": []float32{}}
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
		tick := time.Now()
		event.Process.run()
		duration := time.Since(tick) / time.Nanosecond
		event_type := ""
		if _, ok := event.Process.Message["photon"]; ok {
			t.Type_counter["qc"]++
			t.Type_timer["qc"] += float32(duration) / 1e9
			event_type = "qc"
		} else if _, ok := event.Process.Message["message"]; ok {
			t.Type_counter["cc"]++
			t.Type_timer["cc"] += float32(duration) / 1e9
			event_type = "cc"
		} else if _, ok := event.Process.Message["state"]; ok {
			t.Type_counter["qc"]++
			t.Type_timer["qc"] += float32(duration) / 1e9
			event_type = "qc"
		} else {
			t.Type_counter["other"]++
			t.Type_timer["other"] += float32(duration) / 1e9
			event_type = "other"
			//fmt.Println(event.Process.Message)
		}
		if event_type == "qc" {
			sample := rand.Float32()
			if sample < 0.0001 {
				t.Type_sample["qc"] = append(t.Type_sample["qc"], float32(duration))
			}
		} else if event_type == "cc" {
			sample := rand.Float32()
			if sample < 0.1 {
				t.Type_sample["cc"] = append(t.Type_sample["cc"], float32(duration))
			}
		} else {
			sample := rand.Float32()
			if sample < 0.1 {
				t.Type_sample["other"] = append(t.Type_sample["other"], float32(duration))
			}
		}
	}
}

func (t *Timeline) cleanEvenbuffer() {
	t.eventBuffer.clean()
}

func (t *Timeline) run(br *Barrier, wg *sync.WaitGroup) {
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
	wg.Done()
}
