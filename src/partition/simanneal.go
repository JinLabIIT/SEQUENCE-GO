package partition

import (
	"math"
	"math/rand"
	"sort"
)

type EdgeAttribute struct {
	Weight    int64 // edge does not exist if Weight == 0
	Ratio     float64
	LookAhead int64
}

type PartitionState struct {
	graph     [][]EdgeAttribute
	State     []map[int]bool
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
		panic("State is empty")
	}
	if len(graph) != len(graph[0]) {
		panic("size of x and y dimensions are different")
	}

	return &PartitionState{graph, state, vMoveProb, rng}
}

func (self *PartitionState) Copy() interface{} {
	state := make([]map[int]bool, len(self.State))
	for i := range state {
		state[i] = make(map[int]bool)
		for key, value := range self.State[i] {
			state[i][key] = value
		}
	}
	return &PartitionState{
		graph:     self.graph,
		State:     state,
		vMoveProb: self.vMoveProb,
		rng:       self.rng,
	}
}

func (self *PartitionState) Move() {
	if self.rng.Float64() < self.vMoveProb {
		// move single node to another subset
		var srcIndex int
		for {
			srcIndex = self.rng.Intn(len(self.State))
			if len(self.State[srcIndex]) > 0 {
				break
			}
		}
		eIndex := self.rng.Intn(len(self.State[srcIndex]))
		dstIndex := self.rng.Intn(len(self.State))
		element := getElementByIndex(self.State[srcIndex], eIndex)
		delete(self.State[srcIndex], element)
		_, exist := self.State[dstIndex][element]

		if exist {
			panic("node exists in subset")
		}

		self.State[dstIndex][element] = true
	} else {
		// exchange two nodes from different subsets
	}
}

func (self *PartitionState) Energy() float64 {
	lookahead := self.getLookAhead()
	return (self.getMaxExeTime(lookahead) + self.getMaxMergeTime(lookahead)) * (1 / lookahead)
}

func (self *PartitionState) getMaxExeTime(lookahead float64) float64 {
	maxWeight := float64(0)
	for i := range self.State {
		totalWeight := float64(0)
		for nodeId := range self.State[i] {
			// from nodeId to others
			for _, edge := range self.graph[nodeId] {
				totalWeight += float64(edge.Weight)
			}
			// from others to nodeId
			for j := range self.graph {
				totalWeight += float64(self.graph[j][nodeId].Weight) * self.graph[j][nodeId].Ratio
			}
		}
		maxWeight = maxfloat64(maxWeight, totalWeight)
	}

	return getExeTime(maxWeight, lookahead)
}

func (self *PartitionState) getMaxMergeTime(lookahead float64) float64 {
	maxMergeTime := float64(0)
	for _, state := range self.State {
		totalWeight := float64(0)
		ids := make(map[int]bool, 0)
		for nodeId := range state {
			// from nodeId to others
			for _, edge := range self.graph[nodeId] {
				totalWeight += float64(edge.Weight)
			}
			// from others to nodeId
			for j := range self.graph {
				totalWeight += float64(self.graph[j][nodeId].Weight) * self.graph[j][nodeId].Ratio
				_, exist := state[j]
				if self.graph[j][nodeId].Weight > 0 && !exist {
					ids[j] = true
				}
			}
		}
		outNum := 0
		for _, st := range self.State {
			for id := range ids {
				_, exist := st[id]
				if exist {
					outNum += 1
					break
				}
			}
		}
		mergeTime := getMergeTime(totalWeight, outNum, lookahead)
		maxMergeTime = maxfloat64(maxMergeTime, mergeTime)
	}
	return maxMergeTime
}

func (self *PartitionState) getLookAhead() float64 {
	var lookahead int64
	lookahead = math.MaxInt64

	for i := 0; i < len(self.graph); i++ {
		for j := 0; j < len(self.graph); j++ {
			if self.graph[i][j].Weight != 0 {
				// exist edge
				if !self.twoNodesBelongSameSet(i, j) {
					lookahead = min(lookahead, self.graph[i][j].LookAhead)
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
	for _, subset := range self.State {
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

func getExeTime(weight float64, lookahead float64) float64 {
	k := 1.0
	b := 0.0
	return k*weight*lookahead + b
}

func getMergeTime(weight float64, outNum int, lookahead float64) float64 {
	k := 1.0
	b := 0.0
	return k*math.Ceil(math.Log2(float64(outNum+1)))*weight*lookahead + b
}

func min(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func maxfloat64(a, b float64) float64 {
	if a > b {
		return a
	} else {
		return b
	}
}
