package kernel

import (
	"container/heap"
)

type EventList struct {
	events PriorityQueue
}

//return the size of the eventlist
func (eventlist *EventList) size() int {
	return eventlist.events.Len()
}

//return the first element of the eventlist
func (eventlist *EventList) top() *Event {
	return eventlist.events[0]
}

//push
func (eventlist *EventList) push(event *Event) {
	heap.Push(&eventlist.events, event)
}

//pop
func (eventlist *EventList) pop() *Event {
	return heap.Pop(&eventlist.events).(*Event)
}

func (EventList *EventList) merge(new EventList) {
	a := EventList.events.Len()
	b := new.events.Len()
	n := a + b
	for i := 0; i < new.events.Len(); i++ {
		EventList.events = append(EventList.events, new.events[i])
	}
	for i := int(n/2) - 1; i >= 0; i-- {
		EventList.minHeapify(n, i)
	}
}

func (EventList *EventList) minHeapify(n, i int) {
	if i >= n {
		return
	}
	l := i*2 + 1
	r := i*2 + 2
	var min int
	if l < n && EventList.events.Less(l, i) {
		min = l
	} else {
		min = i
	}
	if r < n && EventList.events.Less(r, min) {
		min = r
	}
	if min != i {
		EventList.events.Swap(min, i)
		EventList.minHeapify(n, min)
	}
}

type PriorityQueue []*Event

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Time < pq[j].Time || (pq[i].Time == pq[j].Time && pq[i].Priority < pq[j].Priority)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	event := x.(*Event)
	*pq = append(*pq, event)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	event := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return event
}
