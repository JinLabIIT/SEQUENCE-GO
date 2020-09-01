package quantum

import (
	"math"
)

type Photon struct {
	//quantumState     []complex128
	encodingType     map[string]interface{} // temp
	firstState       complex128
	secondState      complex128
}

func (photon *Photon) _init() {
	if photon.encodingType == nil {
		photon.encodingType = polarization()
	}
	if photon.firstState == 0 && photon.secondState == 0{ // nil
		//photon.quantumState = []complex128{complex128(1), complex128(0)}
		photon.firstState = complex128(1)
		photon.secondState = complex128(0)
	}
}

func (photon *Photon) randomNoise(noise float64) {
	angle := noise * 2 * math.Pi
	//photon.quantumState = []complex128{complex(math.Cos(angle), 0), complex(math.Sin(angle), 0)}
	photon.firstState = complex(math.Cos(angle), 0)
	photon.secondState = complex(math.Sin(angle), 0)
}

func (photon *Photon) setState(state []complex128) {
	photon.firstState = state[0]
	photon.secondState = state[1]
}

func (photon *Photon) measure(basis *[][]complex128, prob float64) int {
	// only work for BB84
	//state := oneToTwo(&photon.quantumState) // 1-D array to 2-D array
	//fmt.Println()
	state := oneToTwo(&[]complex128{photon.firstState,photon.secondState})
	u := &(*basis)[0]
	v := &(*basis)[1]
	// measurement operator
	M0 := outer(arrayConj(u), u)
	M1 := outer(arrayConj(v), v)

	var projector0 *[][]complex128
	var projector1 *[][]complex128
	projector0 = kron(&[][]complex128{{1}}, M0)
	projector1 = kron(&[][]complex128{{1}}, M1)

	//tmp := matMul(state.conj().transpose(), projector0.conj().transpose())
	tmp := matMul(transpose(conj(state)),transpose(conj(projector0)))
	tmp = matMul(tmp, projector0)
	tmp = matMul(tmp, state)
	// tmp = state.conj().transpose() @ projector0.conj().transpose() @ projector0 @ state
	prob0 := real((*tmp)[0][0])
	result := 0
	if prob > prob0 { // given by the function
		result = 1
	}

	var newState []complex128
	if result == 1 {
		newState = divide(matMul(projector1, state), math.Sqrt(1-prob0))
	} else {
		newState = divide(matMul(projector0, state), math.Sqrt(prob0))
	}
	photon.firstState = newState[0]
	photon.secondState = newState[1]

	return result
}
