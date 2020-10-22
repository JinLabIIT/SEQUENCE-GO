package kernel

import (
	"math"
	"sort"
)

type Rung struct {
	bucketWidth uint64
	buckets     [][]*Event
	rCur        uint64
	rStart      uint64
}

func (r *Rung) initialize(bucketWidth, rStart uint64) {
	if bucketWidth < 1 {
		r.bucketWidth = 1
	} else {
		r.bucketWidth = bucketWidth
	}

	r.rCur = rStart
	r.rStart = rStart
}

func (r *Rung) nextBucket() []*Event {
	index := int((r.rCur - r.rStart) / r.bucketWidth)
	bw := r.bucketWidth
	rCur := r.rCur
	for ; index < len(r.buckets) && len(r.buckets[index]) == 0; index++ {
		rCur += bw
	}
	r.rCur = bw + rCur
	if index == len(r.buckets) {
		r.bucketWidth = 0
		return []*Event{}
	} else {
		bucket := r.buckets[index]
		r.buckets[index] = []*Event{}
		return bucket
	}
}

func (r *Rung) load(events []*Event) {
	if r.bucketWidth == 0 {
		panic("you may forget initialize Rung before load linked list")
	}

	for i := 0; i < len(events); i++ {
		index := int((events[i].Time - r.rStart) / r.bucketWidth)
		for len(r.buckets) <= index {
			r.buckets = append(r.buckets, []*Event{})
		}
		r.buckets[index] = append(r.buckets[index], events[i])
	}
}

type LadderQ struct {
	top        []*Event
	top_max_ts uint64
	top_min_ts uint64
	top_start  uint64

	ladder []*Rung
	rung_n int

	bottom []*Event
	thres  int

	counter int
}

func (lq *LadderQ) Push(event *Event) {
	lq.counter++

	if event.Time >= lq.top_start {
		lq.top = append(lq.top, event)
		lq.top_max_ts = max64(lq.top_max_ts, event.Time)
		lq.top_min_ts = min64(lq.top_min_ts, event.Time)
	} else {
		insertFlag := false
		for i := 0; i < lq.rung_n; i++ {
			if event.Time >= lq.ladder[i].rCur {
				lq.ladder[i].load([]*Event{event})
				insertFlag = true
				break
			}
		}

		if !insertFlag {
			index := 0
			for ; index < len(lq.bottom) && lq.bottom[index].Time < event.Time; index++ {
			}
			lq.bottom = insert(lq.bottom, event, index)

			if len(lq.bottom) >= lq.thres && lq.rung_n < len(lq.ladder) {
				bucketWidth := lq.ladder[lq.rung_n-1].bucketWidth / uint64(len(lq.bottom))
				rStart := lq.bottom[0].Time
				lq.ladder[lq.rung_n].initialize(bucketWidth, rStart)
				lq.ladder[lq.rung_n].load(lq.bottom)
				lq.bottom = []*Event{}
				lq.rung_n++
			}
		}
	}
}

func (lq *LadderQ) Pop() *Event {
	if lq.counter == 0 {
		panic("ladder queue is empty")
	}
	lq.counter--

	if len(lq.bottom) > 0 {
		return lq.popBottom()
	}

	lq.generate_bottom()

	if len(lq.bottom) > 0 {
		return lq.popBottom()
	} else {
		rung := lq.ladder[0]
		bucketWidth := (lq.top_max_ts - lq.top_min_ts) / uint64(len(lq.top))
		rung.initialize(bucketWidth, lq.top_min_ts)
		rung.load(lq.top)

		lq.top_min_ts = 0
		lq.top_start = lq.top_max_ts + bucketWidth
		lq.top_max_ts = 0
		lq.rung_n++
		lq.top = []*Event{}

		lq.generate_bottom()
		return lq.popBottom()
	}

}

func (lq *LadderQ) popBottom() *Event {
	e := lq.bottom[0]
	lq.bottom = lq.bottom[1:]
	return e
}

type KeyFunc []*Event

func (events KeyFunc) Less(i, j int) bool {
	return events[i].Time < events[j].Time || (events[i].Time == events[j].Time && events[i].Priority < events[j].Priority)
}

func (events KeyFunc) Len() int {
	return len(events)
}

func (events KeyFunc) Swap(i, j int) {
	events[i], events[j] = events[j], events[i]
}

func (lq *LadderQ) generate_bottom() {
	for lq.rung_n > 0 {
		rung := lq.ladder[lq.rung_n-1]
		bucket := rung.nextBucket()
		if len(bucket) == 0 {
			lq.rung_n -= 1
		} else {
			if len(bucket) < lq.thres || lq.rung_n == len(lq.ladder) {
				sort.Sort(KeyFunc(bucket))
				lq.bottom = bucket
				return
			} else {
				nextRung := lq.ladder[lq.rung_n]
				bucketWidth := rung.bucketWidth / uint64(len(bucket))
				rStart := rung.rCur - rung.bucketWidth
				nextRung.initialize(bucketWidth, rStart)
				nextRung.load(bucket)
				lq.rung_n++
			}
		}
	}
}

func (lq *LadderQ) Empty() bool {
	return lq.counter == 0
}

func (lq *LadderQ) Size() int {
	return lq.counter
}

func (lq *LadderQ) Top() *Event {
	if len(lq.bottom) > 0 {
		return lq.bottom[0]
	} else {
		event := lq.Pop()
		lq.bottom = append([]*Event{event}, lq.bottom...)
		lq.counter++
		return event
	}
}

func NewLadderQ(thres, max_rung int) *LadderQ {
	ladder := []*Rung{}
	for i := 0; i < max_rung; i++ {
		ladder = append(ladder, &Rung{
			bucketWidth: 0,
			buckets:     [][]*Event{},
			rCur:        0,
			rStart:      0,
		})
	}
	lq := LadderQ{
		top:        []*Event{},
		top_max_ts: 0,
		top_min_ts: math.MaxUint64,
		top_start:  0,
		ladder:     ladder,
		rung_n:     0,
		bottom:     []*Event{},
		thres:      thres,
		counter:    0,
	}

	return &lq
}

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func insert(events []*Event, element *Event, index int) []*Event {
	return append(events[:index], append([]*Event{element}, events[index:]...)...)
}
