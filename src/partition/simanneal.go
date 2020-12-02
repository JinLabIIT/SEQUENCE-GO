package partition

import (
	"math"
	"math/rand"
	"sort"
)

type EdgeAttribute struct {
	Weight    float64 // edge does not exist if Weight == 0
	Ratio     float64
	LookAhead int64
}

type PartitionState struct {
	graph     [][]EdgeAttribute
	State     []map[int]bool
	vMoveProb float64
	simTime   float64
	rng       *rand.Rand
}

func NewPartitionState(graph [][]EdgeAttribute, state []map[int]bool, vMoveProb, simTime float64, seed int64) *PartitionState {
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

	return &PartitionState{graph, state, vMoveProb, simTime, rng}
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
		simTime:   self.simTime,
		rng:       self.rng,
	}
}

func (self *PartitionState) Move() {
	if self.simTime == 0 {
		panic("sim time is 0")
	}
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
		validSets := []int{}
		for i, state := range self.State {
			if len(state) > 0 {
				validSets = append(validSets, i)
			}
		}
		if len(validSets) < 2 {
			return
		}
		setId1 := validSets[self.rng.Intn(len(validSets))]
		validSets = append(validSets[:setId1], validSets[setId1+1:]...)
		setId2 := validSets[self.rng.Intn(len(validSets))]
		eIndex1 := self.rng.Intn(len(self.State[setId1]))
		eIndex2 := self.rng.Intn(len(self.State[setId2]))
		key1 := getElementByIndex(self.State[setId1], eIndex1)
		key2 := getElementByIndex(self.State[setId2], eIndex2)
		//swap
		delete(self.State[setId1], key1)
		delete(self.State[setId2], key2)

		self.State[setId1][key2] = true
		self.State[setId2][key1] = true
	}
}

func (self *PartitionState) Energy() float64 {
	lookahead := self.GetLookAhead()
	//fmt.Println(self.getComputeTime(), self.getCommTime(), self.GetSyncTime(lookahead))
	return self.getComputeTime() + self.getCommTime() + self.GetSyncTime(lookahead)
}

func (self *PartitionState) getComputeTime() float64 {
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

	return getExeTime(maxWeight, len(self.State)) * 1e-12 * self.simTime
}

func (self *PartitionState) GetEventNum(lookahead float64) int64 {
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
	return int64(maxWeight * lookahead)
}

func (self *PartitionState) getCommTime() float64 {
	maxEvent := float64(0)
	for _, state := range self.State {
		totalWeight := float64(0)
		for nodeId := range state {
			// from others to nodeId
			for j := range self.graph {
				totalWeight += float64(self.graph[j][nodeId].Weight) * self.graph[j][nodeId].Ratio
			}
		}

		maxEvent = maxfloat64(maxEvent, totalWeight)
	}
	return getMergeTime(maxEvent, len(self.State)) * 1e-12 * self.simTime
}

func (self *PartitionState) GetSyncTime(lookahead float64) float64 {
	a0 := -2504085.791
	a1 := 86181.449
	a2 := 6057.241
	syncCounter := self.simTime / lookahead
	return a0 + a1*syncCounter + a2*syncCounter*float64(len(self.State))
}

func (self *PartitionState) GetLookAhead() float64 {
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

func getExeTime(event_num float64, threadNum int) float64 {
	// nano second
	//k := 6881.0
	//b := 4536000.0
	//return k*weight*lookahead + b
	a0 := 2.312352e+09
	a1 := 3.754279e+03
	a2 := 1.773645e+02
	a3 := -8.931743e+07
	return a0 + a1*event_num + a2*event_num*float64(threadNum) + a3*float64(threadNum)
}

func getMergeTime(event_num float64, threadNum int) float64 {
	// nano second
	a0 := 5.304552e+05
	a1 := 2.543406e+02
	a2 := 7.172598e-01
	a3 := 6.303950e+04
	return a0 + a1*event_num + a2*event_num*float64(threadNum) + a3*float64(threadNum)
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
