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

func (EventList *EventList) merge(new EventList){
	a := EventList.events.Len()
	b := new.events.Len()
	n := a+b
	for i := 0; i < new.events.Len(); i++{
		EventList.events = append(EventList.events, new.events[i])
	}
	for i := int(n / 2) - 1; i >= 0; i--{
		EventList.minHeapify(n, i)
	}
}

func (EventList *EventList) minHeapify(n,i int){
	if i >= n {
		return
	}
	l := i * 2 + 1
	r := i * 2 + 2
	var min int
	if l < n && EventList.events.Less(l,i) {
		min = l
	}else{
		min = i
	}
	if r < n && EventList.events.Less(r,min){
		min = r
	}
	if min != i {
		EventList.events.Swap(min,i)
		EventList.minHeapify(n,min)
	}
}

type PriorityQueue []*Event

func (evenlist PriorityQueue) Len() int { return len(evenlist) }

func (evenlist PriorityQueue) Less(i, j int) bool {
	return evenlist[i].time < evenlist[j].time || (evenlist[i].time == evenlist[j].time && evenlist[i].priority < evenlist[j].priority)
}

func (evenlist PriorityQueue) Swap(i, j int) {
	evenlist[i], evenlist[j] = evenlist[j], evenlist[i]
}

func (evenlist *PriorityQueue) Push(x interface{}) {
	event := x.(*Event)
	*evenlist = append(*evenlist, event)
}

func (evenlist *PriorityQueue) Pop() interface{} {
	old := *evenlist
	n := len(old)
	event := old[n-1]
	old[n-1] = nil // avoid memory leak
	*evenlist = old[0 : n-1]
	return event
}
