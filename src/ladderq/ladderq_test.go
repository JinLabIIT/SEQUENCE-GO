package ladderq

import (
	"golang.org/x/exp/errors/fmt"
	"golang.org/x/exp/rand"
	"kernel"
	"reflect"
	"testing"
	"time"
)

func TestRung_initialize(t *testing.T) {
	type fields struct {
		bucketWidth uint64
		buckets     [][]*kernel.Event
		rCur        uint64
		rStart      uint64
	}
	type args struct {
		bucketWidth uint64
		rStart      uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"test1", fields{1, [][]*kernel.Event{}, 0, 0}, args{10, 100}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rung{
				bucketWidth: tt.fields.bucketWidth,
				buckets:     tt.fields.buckets,
				rCur:        tt.fields.rCur,
				rStart:      tt.fields.rStart,
			}
			if r.bucketWidth != tt.fields.bucketWidth {
				t.Error("initial bucket width is differnt")
			}

			r.initialize(10, 20)
			if r.bucketWidth != 10 || r.rStart != 20 || r.rCur != 20 {
				t.Error("initialize function error", r.bucketWidth, r.rStart, r.rCur)
			}
		})
	}
}

func TestRung_nextBucket(t *testing.T) {
	type fields struct {
		bucketWidth uint64
		buckets     [][]*kernel.Event
		rCur        uint64
		rStart      uint64
	}

	event := kernel.Event{
		1,
		0,
		nil,
	}
	tests := []struct {
		name      string
		fields    fields
		want      []*kernel.Event
		expectCur uint64
	}{
		{"no bucket", fields{
			bucketWidth: 1,
			buckets:     [][]*kernel.Event{},
			rCur:        0,
			rStart:      0,
		}, []*kernel.Event{}, 1},
		{"has bucket", fields{
			bucketWidth: 1,
			buckets:     [][]*kernel.Event{{&event}},
			rCur:        0,
			rStart:      0,
		}, []*kernel.Event{&event}, 1},
		{"has event in the second bucket", fields{
			bucketWidth: 1,
			buckets:     [][]*kernel.Event{{}, {&event}},
			rCur:        1,
			rStart:      0,
		}, []*kernel.Event{&event}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rung{
				bucketWidth: tt.fields.bucketWidth,
				buckets:     tt.fields.buckets,
				rCur:        tt.fields.rCur,
				rStart:      tt.fields.rStart,
			}
			if got := r.nextBucket(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nextBucket() = %v, want %v", got, tt.want)
			}

			if r.rCur != tt.expectCur {
				t.Error("rCur is wrong")
			}
		})
	}
}

func TestRung_load(t *testing.T) {
	type fields struct {
		bucketWidth uint64
		buckets     [][]*kernel.Event
		rCur        uint64
		rStart      uint64
	}
	type args struct {
		events []*kernel.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"buckets are created", fields{
			bucketWidth: 1,
			buckets:     [][]*kernel.Event{{}, {}, {}, {}, {}},
			rCur:        0,
			rStart:      0,
		}, args{[]*kernel.Event{
			{
				Time: 0,
			},
			{
				Time: 0,
			},
			{
				Time: 1,
			},
			{
				Time: 4,
			},
			{
				Time: 3,
			},
		}}},

		{"buckets are not created", fields{
			bucketWidth: 1,
			buckets:     [][]*kernel.Event{},
			rCur:        0,
			rStart:      0,
		}, args{[]*kernel.Event{
			{
				Time: 0,
			},
			{
				Time: 0,
			},
			{
				Time: 1,
			},
			{
				Time: 4,
			},
			{
				Time: 3,
			},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rung{
				bucketWidth: tt.fields.bucketWidth,
				buckets:     tt.fields.buckets,
				rCur:        tt.fields.rCur,
				rStart:      tt.fields.rStart,
			}
			r.load(tt.args.events)
			counter := len(tt.args.events)
			for i := 0; i < len(r.buckets); i++ {
				for j := 0; j < len(r.buckets[i]); j++ {
					if r.buckets[i][j].Time != uint64(i) {
						t.Error("event is putted into wrong bucket")
					}
					counter--
				}
			}
			if counter != 0 {
				t.Error("Some events are miss or duplicated.", counter)
			}
		})
	}
}

func TestNewLadderQ(t *testing.T) {
	type args struct {
		thres    int
		max_rung int
	}
	tests := []struct {
		name string
		args args
	}{
		{"constructor of ladder q", args{
			thres:    10,
			max_rung: 5,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLadderQ(tt.args.thres, tt.args.max_rung)
			if got.thres != tt.args.thres {
				t.Error("threshold fails in initialization")
			}

			if len(got.ladder) != tt.args.max_rung {
				t.Error("rungs are initialized in wrong number", len(got.ladder), tt.args.max_rung)
			}
		})
	}
}

func TestLadderQ(t *testing.T) {
	lq := NewLadderQ(50, 5)
	ts := []uint64{9, 8, 7, 6, 5, 4, 3}
	for i := 0; i < len(ts); i++ {
		lq.Push(&kernel.Event{
			Time: ts[i],
		})
	}
	if len(lq.top) != len(ts) {
		t.Error("Push is wrong")
	}

	events := []*kernel.Event{}

	for !lq.Empty() {
		events = append(events, lq.Pop())
	}

	for i := 1; i < len(events); i++ {
		if events[i].Time < events[i-1].Time {
			t.Error("Pop is wrong")
		}
	}

}

func TestSpeed(t *testing.T) {
	tick := time.Now()
	queueSize := 10000
	round := 100000000
	lq := NewLadderQ(50, 8)
	for i := 0; i < queueSize; i++ {
		lq.Push(&kernel.Event{
			Time: uint64(rand.Intn(10000)),
		})
	}

	last_ts := uint64(0)
	for i := 0; i < round; i++ {
		event := lq.Pop()
		if last_ts > event.Time {
			t.Error("Pop wrong")
		}
		last_ts = event.Time
		event.Time += uint64(rand.Intn(10000))
		lq.Push(event)
	}
	fmt.Println(time.Since(tick))
}
