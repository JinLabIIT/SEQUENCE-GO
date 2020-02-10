package partition

import (
	"reflect"
	"testing"
)

func TestPartitionState_Copy(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	ea := EdgeAttribute{}
	graph := [][]EdgeAttribute{{ea, ea, ea}, {ea, ea, ea}, {ea, ea, ea}}
	state := []map[int]bool{{1: true, 5: true}, {2: true, 6: true}, {3: true, 4: true}}
	vMoveProb := 1.0
	seed := int64(1)
	args := fields{
		graph:     graph,
		state:     state,
		vMoveProb: vMoveProb,
		seed:      seed,
	}

	tests := []struct {
		name   string
		fields fields
	}{

		{name: "test1", fields: args},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			got := self.Copy().(*PartitionState)
			self.State[0][-1] = true
			if len(self.State[0]) == len(got.State[0]) {
				t.Errorf("Copy() = %v", got)
			}
		})
	}
}

func TestPartitionState_Move(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	ea := EdgeAttribute{}
	tests := []struct {
		name   string
		fields fields
		want   []map[int]bool
	}{
		{"vertexMove1", fields{
			graph:     [][]EdgeAttribute{{ea, ea, ea}, {ea, ea, ea}, {ea, ea, ea}},
			state:     []map[int]bool{{1: true, 5: true}, {2: true, 6: true}, {3: true, 4: true}},
			vMoveProb: 1.0,
			seed:      int64(1),
		}, []map[int]bool{{1: true}, {2: true, 6: true}, {3: true, 4: true, 5: true}}},
		{"vertexMove2", fields{
			graph:     [][]EdgeAttribute{{ea, ea, ea}, {ea, ea, ea}, {ea, ea, ea}},
			state:     []map[int]bool{{1: true}, {}, {}},
			vMoveProb: 1.0,
			seed:      int64(1),
		}, []map[int]bool{{}, {}, {1: true}}},
		{"vertexMove3", fields{
			graph:     [][]EdgeAttribute{{ea, ea, ea}, {ea, ea, ea}, {ea, ea, ea}},
			state:     []map[int]bool{{}, {}, {1: true}},
			vMoveProb: 1.0,
			seed:      int64(1),
		}, []map[int]bool{{}, {1: true}, {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			self.Move()
			reflect.DeepEqual(self.State, tt.want)
		})
	}
}

func Test_getElementByIndex(t *testing.T) {
	type args struct {
		targetSet map[int]bool
		index     int
	}
	targetSet := map[int]bool{16: true, 8: true, 4: true, 2: true, 1: true}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test1", args{targetSet, 0}, 1},
		{"test2", args{targetSet, 1}, 2},
		{"test3", args{targetSet, 2}, 4},
		{"test4", args{targetSet, 3}, 8},
		{"test5", args{targetSet, 4}, 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getElementByIndex(tt.args.targetSet, tt.args.index); got != tt.want {
				t.Errorf("getElementByIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionState_twoNodesBelongSameSet(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	type args struct {
		id1 int
		id2 int
	}

	graph := make([][]EdgeAttribute, 1)
	graph[0] = make([]EdgeAttribute, 1)
	state := []map[int]bool{{1: true, 2: true}, {3: true, 4: true}, {}, {7: true, 8: true}}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"test1",
			fields{
				graph:     graph,
				state:     state,
				vMoveProb: 1,
				seed:      1,
			},
			args{
				id1: 1,
				id2: 2,
			},
			true},
		{"test2",
			fields{
				graph:     graph,
				state:     state,
				vMoveProb: 1,
				seed:      1,
			},
			args{
				id1: 1,
				id2: 3,
			},
			false},
		{"test3",
			fields{
				graph:     graph,
				state:     state,
				vMoveProb: 1,
				seed:      1,
			},
			args{
				id1: 1,
				id2: 7,
			},
			false},
		{"test4",
			fields{
				graph:     graph,
				state:     state,
				vMoveProb: 1,
				seed:      1,
			},
			args{
				id1: 7,
				id2: 8,
			},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)

			if got := self.twoNodesBelongSameSet(tt.args.id1, tt.args.id2); got != tt.want {
				t.Errorf("twoNodesBelongSameSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionState_getLookAhead(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	graph := make([][]EdgeAttribute, 0)
	GRAPHSIZE := 4
	for i := 0; i < GRAPHSIZE; i++ {
		line := make([]EdgeAttribute, GRAPHSIZE)
		graph = append(graph, line)
	}
	graph[0][1] = EdgeAttribute{10, 1, 10}
	graph[1][2] = EdgeAttribute{20, 1, 20}
	graph[2][3] = EdgeAttribute{30, 1, 30}
	graph[3][0] = EdgeAttribute{40, 1, 40}

	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"test1", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true}, {2: true, 3: true}},
			vMoveProb: 1,
			seed:      1,
		}, 20},
		{"test2", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 3: true}, {1: true, 2: true}},
			vMoveProb: 1,
			seed:      1,
		}, 10},
		{"test3", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 3: true}, {2: true}},
			vMoveProb: 1,
			seed:      1,
		}, 20},
		{"test4", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 2: true}, {3: true}},
			vMoveProb: 1,
			seed:      1,
		}, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			if got := self.getLookAhead(); got != tt.want {
				t.Errorf("getLookAhead() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionState_getMaxExeTime(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	type args struct {
		lookahead float64
	}
	graph := make([][]EdgeAttribute, 0)
	GRAPHSIZE := 4
	for i := 0; i < GRAPHSIZE; i++ {
		line := make([]EdgeAttribute, GRAPHSIZE)
		graph = append(graph, line)
	}
	graph[0][1] = EdgeAttribute{10, 1, 10}
	graph[1][2] = EdgeAttribute{20, 1, 20}
	graph[2][3] = EdgeAttribute{30, 1, 30}
	graph[3][0] = EdgeAttribute{40, 1, 40}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{"test1", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true}, {2: true, 3: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{20}, 2400},
		{"test2", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 3: true}, {1: true, 2: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{10}, 1200},
		{"test3", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 3: true}, {2: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{20}, 3000},
		{"test4", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 2: true}, {3: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{30}, 3900},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			if got := self.getMaxExeTime(tt.args.lookahead); got != tt.want {
				t.Errorf("getMaxExeTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionState_getMaxMergeTime(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	type args struct {
		lookahead float64
	}
	graph := make([][]EdgeAttribute, 0)
	GRAPHSIZE := 4
	for i := 0; i < GRAPHSIZE; i++ {
		line := make([]EdgeAttribute, GRAPHSIZE)
		graph = append(graph, line)
	}
	graph[0][1] = EdgeAttribute{10, 1, 10}
	graph[1][2] = EdgeAttribute{20, 1, 20}
	graph[2][3] = EdgeAttribute{30, 1, 30}
	graph[3][0] = EdgeAttribute{40, 1, 40}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{"test1", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true}, {2: true, 3: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{20}, 2400},
		{"test2", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 3: true}, {1: true, 2: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{10}, 1200},
		{"test3", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 3: true}, {2: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{20}, 3000},
		{"test4", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 2: true}, {3: true}},
			vMoveProb: 1,
			seed:      1,
		}, args{30}, 3900},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			if got := self.getMaxMergeTime(tt.args.lookahead); got != tt.want {
				t.Errorf("getMaxMergeTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionState_Energy(t *testing.T) {
	type fields struct {
		graph     [][]EdgeAttribute
		state     []map[int]bool
		vMoveProb float64
		seed      int64
	}
	graph := make([][]EdgeAttribute, 0)
	GRAPHSIZE := 4
	for i := 0; i < GRAPHSIZE; i++ {
		line := make([]EdgeAttribute, GRAPHSIZE)
		graph = append(graph, line)
	}
	graph[0][1] = EdgeAttribute{10, 1, 10}
	graph[1][2] = EdgeAttribute{20, 1, 20}
	graph[2][3] = EdgeAttribute{30, 1, 30}
	graph[3][0] = EdgeAttribute{40, 1, 40}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"test1", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true}, {2: true, 3: true}},
			vMoveProb: 1,
			seed:      1,
		}, 240},
		{"test2", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 3: true}, {1: true, 2: true}},
			vMoveProb: 1,
			seed:      1,
		}, 240},
		{"test3", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 3: true}, {2: true}},
			vMoveProb: 1,
			seed:      1,
		}, 300},
		{"test4", fields{
			graph:     graph,
			state:     []map[int]bool{{0: true, 1: true, 2: true}, {3: true}},
			vMoveProb: 1,
			seed:      1,
		}, 260},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewPartitionState(tt.fields.graph, tt.fields.state, tt.fields.vMoveProb, tt.fields.seed)
			if got := self.Energy(); got != tt.want {
				t.Errorf("Energy() = %v, want %v", got, tt.want)
			}
		})
	}
}
