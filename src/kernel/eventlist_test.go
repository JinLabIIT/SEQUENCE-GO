package kernel

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

func TestPriorityQueue_Len(t *testing.T) {
	pq1 := PriorityQueue{}

	pq2 := PriorityQueue{}
	pq2.Push(&Event{})

	pq3 := PriorityQueue{}
	pq3.Push(&Event{})
	pq3.Pop()

	tests := []struct {
		name string
		pq   PriorityQueue
		want int
	}{
		{"empty pq", pq1, 0},
		{"push operation", pq2, 1},
		{"pop operation", pq3, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pq.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityQueue_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	pq := make([]*Event, 3)
	pq[0] = &Event{time: 2, priority: 1}
	pq[1] = &Event{time: 1, priority: 2}
	pq[2] = &Event{time: 2, priority: 2}
	tests := []struct {
		name string
		pq   PriorityQueue
		args args
		want bool
	}{
		{"test1", pq, args{0, 1}, false},
		{"test2", pq, args{0, 2}, true},
		{"test3", pq, args{1, 2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pq.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventList_push(t *testing.T) {
	type fields struct {
		events PriorityQueue
	}
	type args struct {
		event *Event
	}
	pq := make(PriorityQueue, 0)
	elist1 := fields{events: pq}
	event1 := &Event{time: 2, priority: 1}
	args1 := args{event1}
	wants1 := Event{time: 2, priority: 1}

	// pq = make(PriorityQueue, 0)
	eventlist := EventList{}
	eventlist.push(event1)
	elist2 := fields{eventlist.events}
	event2 := &Event{time: 1, priority: 2}
	args2 := args{event2}
	wants2 := Event{time: 1, priority: 2}

	eventlist = EventList{}
	eventlist.push(event1)
	eventlist.push(event2)
	elist3 := fields{eventlist.events}
	event3 := &Event{time: 2, priority: 2}
	args3 := args{event3}
	wants3 := Event{time: 1, priority: 2}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   Event
	}{
		{"test1", elist1, args1, wants1},
		{"test2", elist2, args2, wants2},
		{"test3", elist3, args3, wants3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventlist := &EventList{
				events: tt.fields.events,
			}
			if eventlist.push(tt.args.event); (*eventlist.top()) != tt.want {
				t.Errorf("%s: top() = %v, want %v", tt.name, eventlist.top(), tt.want)
			}
		})
	}
	//test zb 35 33 42 10 14 19 27 44 26 31
	fmt.Println("zb")
	eventlist = EventList{}
	eventlist.push(&Event{time: 35, priority: 0})
	eventlist.push(&Event{time: 33, priority: 0})
	eventlist.push(&Event{time: 42, priority: 0})
	eventlist.push(&Event{time: 10, priority: 0})
	eventlist.push(&Event{time: 14, priority: 0})
	eventlist.push(&Event{time: 19, priority: 0})
	eventlist.push(&Event{time: 27, priority: 0})
	eventlist.push(&Event{time: 44, priority: 0})
	eventlist.push(&Event{time: 26, priority: 0})
	eventlist.push(&Event{time: 31, priority: 0})
	for eventlist.size() > 0 {
		fmt.Print("pop")
		fmt.Println(eventlist.pop())
	}
}

func TestEventList_pop(t *testing.T) {
	type fields struct {
		events PriorityQueue
	}
	type wants []*Event

	eventlist := EventList{}
	event1 := &Event{time: 2, priority: 1}
	event2 := &Event{time: 1, priority: 1}
	event3 := &Event{time: 15, priority: 1}
	event4 := &Event{time: 3, priority: 1}
	event5 := &Event{time: 12, priority: 1}
	eventlist.push(event1)
	eventlist.push(event2)
	eventlist.push(event3)
	eventlist.push(event4)
	eventlist.push(event5)
	fields1 := fields{eventlist.events}

	want1 := make(wants, eventlist.size())
	want1[0] = event2
	want1[1] = event1
	want1[2] = event4
	want1[3] = event5
	want1[4] = event3

	eventlist = EventList{}
	event1 = &Event{priority: 2, time: 1}
	event2 = &Event{priority: 1, time: 1}
	event3 = &Event{priority: 15, time: 1}
	event4 = &Event{priority: 3, time: 1}
	event5 = &Event{priority: 12, time: 1}
	eventlist.push(event1)
	eventlist.push(event2)
	eventlist.push(event3)
	eventlist.push(event4)
	eventlist.push(event5)
	fields2 := fields{eventlist.events}

	want2 := make(wants, eventlist.size())
	want2[0] = event2
	want2[1] = event1
	want2[2] = event4
	want2[3] = event5
	want2[4] = event3

	eventlist = EventList{}
	event1 = &Event{priority: 2, time: 1}
	event2 = &Event{priority: 1, time: 2}
	event3 = &Event{priority: 15, time: 1}
	event4 = &Event{priority: 3, time: 1}
	event5 = &Event{priority: 12, time: 1}
	eventlist.push(event1)
	eventlist.push(event2)
	eventlist.push(event3)
	eventlist.push(event4)
	eventlist.push(event5)
	fields3 := fields{eventlist.events}

	want3 := make(wants, eventlist.size())
	want3[0] = event1
	want3[1] = event4
	want3[2] = event5
	want3[3] = event3
	want3[4] = event2

	tests := []struct {
		name   string
		fields fields
		want   wants
	}{
		{"test1", fields1, want1},
		{"test2", fields2, want2},
		{"test3", fields3, want3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventlist := &EventList{
				events: tt.fields.events,
			}

			for i := 0; eventlist.size() > 0; i++ {
				if got := eventlist.pop(); !reflect.DeepEqual(got, tt.want[i]) {
					t.Errorf("round %d: pop() = %v, want %v", i, got, tt.want[i])
				}
			}

		})
	}
}

func TestEventList_size(t *testing.T) {
	type fields struct {
		events PriorityQueue
	}
	eventlist := EventList{}
	fields1 := fields{eventlist.events}

	eventlist = EventList{}
	n := 5
	for i := 0; i < n; i++ {
		eventlist.push(&Event{})
	}
	fields2 := fields{eventlist.events}

	eventlist = EventList{}
	for i := 0; i < n; i++ {
		eventlist.push(&Event{})
	}
	eventlist.pop()
	fields3 := fields{eventlist.events}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"empty list", fields1, 0},
		{"after push", fields2, n},
		{"after pop", fields3, n - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventlist := &EventList{
				events: tt.fields.events,
			}
			if got := eventlist.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventList_top(t *testing.T) {
	type fields struct {
		events PriorityQueue
	}
	eventlist := EventList{}
	event1 := &Event{priority: 2, time: 1}
	event2 := &Event{priority: 1, time: 2}
	event3 := &Event{priority: 15, time: 1}
	event4 := &Event{priority: 3, time: 1}
	event5 := &Event{priority: 12, time: 1}
	eventlist.push(event1)
	eventlist.push(event2)
	eventlist.push(event3)
	eventlist.push(event4)
	eventlist.push(event5)
	fields1 := fields{eventlist.events}
	tests := []struct {
		name   string
		fields fields
	}{
		{"test1", fields1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventlist := &EventList{
				events: tt.fields.events,
			}
			if got := eventlist.top(); !reflect.DeepEqual(got, eventlist.pop()) {
				t.Errorf("top() != pop() %v", got)
			}
		})
	}
}

func TestEventList_merge(t *testing.T) {
	type fields struct {
		events PriorityQueue
	}
	eventlist := EventList{}
	eventlist2 := EventList{}
	eventlist3 := EventList{}
	a := 474
	b := 632
	for i := 0; i < a; i++ {
		event := &Event{priority: rand.Intn(100), time: uint64(rand.Intn(100))}
		eventlist.push(event)
		eventlist3.push(event)
	}
	for i := 0; i < b; i++ {
		event := &Event{priority: rand.Intn(100), time: uint64(rand.Intn(100))}
		eventlist2.push(event)
		eventlist3.push(event)
	}
	eventlist.merge(eventlist2)

	for eventlist.size() > 0 {
		cc := eventlist.pop()
		bb := eventlist3.pop()
		if cc.time != bb.time || cc.priority != bb.priority {
			fmt.Println(cc.time)
			fmt.Println(bb.time)
			fmt.Println(cc.priority)
			fmt.Println(bb.priority)
			fmt.Println(eventlist.size())
			os.Exit(-1)
		}
	}
}
