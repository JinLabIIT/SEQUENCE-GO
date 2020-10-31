package quantum

import (
	"github.com/golang/groupcache/lru"
	"math"
	"sync"
)

type Photon struct {
	quantumState [2]complex128
	encodingType map[string]interface{} // temp
}

func (photon *Photon) _init() {
	if photon.encodingType == nil {
		photon.encodingType = polarization()
	}
	if photon.quantumState[0] == 0 && photon.quantumState[1] == 0 { // nil
		photon.quantumState = [2]complex128{complex128(1), complex128(0)}
		//photon.firstState = complex128(1)
		//photon.secondState = complex128(0)
	}
}

func (photon *Photon) randomNoise(noise float64) {
	angle := noise * 2 * math.Pi
	photon.quantumState = [2]complex128{complex(math.Cos(angle), 0), complex(math.Sin(angle), 0)}
	//photon.firstState = complex(math.Cos(angle), 0)
	//photon.secondState = complex(math.Sin(angle), 0)
}

func (photon *Photon) setState(state [2]complex128) {
	photon.quantumState = state
}

func (photon *Photon) measure(basis *[2][2]complex128, prob float64) int {
	// only work for BB84
	//state := oneToTwo(&photon.quantumState) // 1-D array to 2-D array
	//fmt.Println()

	state0, state1, prob0 := _meansure(photon.quantumState, *basis)
	result := 0
	if prob > prob0 { // given by the function
		result = 1
	}

	if result == 1 {
		photon.quantumState = state1
	} else {
		photon.quantumState = state0
	}

	return result
}

var cache = lru.New(2000)
var rwMutex = &sync.RWMutex{}

func _meansure(quantumState [2]complex128, basis [2][2]complex128) ([2]complex128, [2]complex128, float64) {
	args := [2]interface{}{quantumState, basis}
	rwMutex.RLock()
	res, exist := cache.Get(args)
	rwMutex.RUnlock()
	if exist {
		result := res.([3]interface{})
		return result[0].([2]complex128), result[1].([2]complex128), result[2].(float64)
	}

	state := oneToTwo(&[]complex128{quantumState[0], quantumState[1]})
	u := &(basis)[0]
	v := &(basis)[1]
	// measurement operator
	M0 := outer(arrayConj(u), &[]complex128{(basis)[0][0], (basis)[0][1]})
	M1 := outer(arrayConj(v), &[]complex128{(basis)[1][0], (basis)[1][1]})

	var projector0 *[][]complex128
	var projector1 *[][]complex128
	projector0 = kron(&[][]complex128{{1}}, M0)
	projector1 = kron(&[][]complex128{{1}}, M1)

	//tmp := matMul(state.conj().transpose(), projector0.conj().transpose())
	tmp := matMul(transpose(conj(state)), transpose(conj(projector0)))
	tmp = matMul(tmp, projector0)
	tmp = matMul(tmp, state)
	// tmp = state.conj().transpose() @ projector0.conj().transpose() @ projector0 @ state
	prob0 := real((*tmp)[0][0])

	var state0 = [2]complex128{}
	var state1 = [2]complex128{}

	if prob0 < 1 {
		copy(state1[:2], divide(matMul(projector1, state), math.Sqrt(1-prob0)))
	}

	if prob0 > 0 {
		copy(state0[:2], divide(matMul(projector0, state), math.Sqrt(prob0)))
	}
	value := [3]interface{}{state0, state1, prob0}
	if cache.Len() < cache.MaxEntries {
		rwMutex.Lock()
		cache.Add(args, value)
		rwMutex.Unlock()
	}
	return state0, state1, prob0
}
