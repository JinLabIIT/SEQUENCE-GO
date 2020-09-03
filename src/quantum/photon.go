package quantum

import (
	"math"
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
	state := oneToTwo(&[]complex128{photon.quantumState[0], photon.quantumState[1]})
	u := &(*basis)[0]
	v := &(*basis)[1]
	// measurement operator
	M0 := outer(arrayConj(u), &[]complex128{(*basis)[0][0], (*basis)[0][1]})
	M1 := outer(arrayConj(v), &[]complex128{(*basis)[1][0], (*basis)[1][1]})

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
	photon.quantumState = [2]complex128{newState[0], newState[1]}
	return result
}
