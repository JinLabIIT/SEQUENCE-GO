package partition

import (
	"math"
	"math/rand"
	"sort"
)

type EdgeAttribute struct {
	load      int64 // edge does not exist if load == 0
	ratio     float64
	lookAhead int64
}

type PartitionState struct {
	graph     [][]EdgeAttribute
	state     []map[int]bool
	vMoveProb float64
	rng       *rand.Rand
}

func NewPartitionState(graph [][]EdgeAttribute, state []map[int]bool, vMoveProb float64, seed int64) *PartitionState {
	rng := rand.New(rand.NewSource(seed))
	flag := true
	for i := range state {
		if len(state[i]) > 0 {
			flag = false
		}
	}

	if flag {
		panic("state is empty")
	}
	if len(graph) != len(graph[0]) {
		panic("size of x and y dimensions are different")
	}

	return &PartitionState{graph, state, vMoveProb, rng}
}

func (self *PartitionState) Copy() interface{} {
	state := make([]map[int]bool, len(self.state))
	for i := range state {
		state[i] = make(map[int]bool)
		for key, value := range self.state[i] {
			state[i][key] = value
		}
	}
	return &PartitionState{
		graph:     self.graph,
		state:     state,
		vMoveProb: self.vMoveProb,
		rng:       self.rng,
	}
}

func (self *PartitionState) Move() {
	//todo: avoid state that all elements are assigned to one subset
	if self.rng.Float64() < self.vMoveProb {
		// move single node to another subset
		var srcIndex int
		for {
			srcIndex = self.rng.Intn(len(self.state))
			if len(self.state[srcIndex]) > 0 {
				break
			}
		}
		eIndex := self.rng.Intn(len(self.state[srcIndex]))
		dstIndex := self.rng.Intn(len(self.state))
		element := getElementByIndex(self.state[srcIndex], eIndex)
		delete(self.state[srcIndex], element)
		_, exist := self.state[dstIndex][element]

		if exist {
			panic("node exists in subset")
		}

		self.state[dstIndex][element] = true
	} else {
		// exchange two nodes from different subsets
	}
}

func (self *PartitionState) Energy() float64 {
	lookahead := self.getLookAhead()
	return (self.getMaxExeTime(lookahead) + self.getMaxMergeTime(lookahead)) * (1 / lookahead)
}

func (self *PartitionState) getMaxExeTime(lookahead float64) float64 {
	return 0
}

func (self *PartitionState) getMaxMergeTime(lookahead float64) float64 {
	return 0
}

func (self *PartitionState) getLookAhead() float64 {
	var lookahead int64
	lookahead = math.MaxInt64

	for i := 0; i < len(self.graph); i++ {
		for j := 0; j < len(self.graph); j++ {
			if self.graph[i][j].load != 0 {
				// exist edge
				if !self.twoNodesBelongSameSet(i, j) {
					lookahead = min(lookahead, self.graph[i][j].lookAhead)
				}
			} else {
				// nil edge
				continue
			}
		}
	}

	return float64(lookahead)
}

func (self *PartitionState) twoNodesBelongSameSet(id1, id2 int) bool {
	for _, subset := range self.state {
		_, exist1 := subset[id1]
		_, exist2 := subset[id2]
		if exist1 && exist2 {
			return true
		}
	}
	return false
}

func getElementByIndex(targetSet map[int]bool, index int) int {
	var keys []int
	for k := range targetSet {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys[index]
}

func min(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}
