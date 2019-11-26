package kernel

import (
	"container/heap"
)

type EventList struct {
	events PriorityQueue
}

func (eventlist *EventList) size() int {
	return eventlist.events.Len()
}

func (eventlist *EventList) top() *Event {
	return eventlist.events[0]
}

func (eventlist *EventList) push(event *Event) {
	heap.Push(&eventlist.events, event)
}

func (eventlist *EventList) pop() *Event {
	return heap.Pop(&eventlist.events).(*Event)
}

func (evenlist *EventList) merge(another *EventList) {
	// TODO
}

type PriorityQueue []*Event

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].time < pq[j].time || (pq[i].time == pq[j].time && pq[i].priority < pq[j].priority)
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
