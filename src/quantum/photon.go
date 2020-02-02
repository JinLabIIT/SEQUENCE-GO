package quantum

import (
	"kernel"
	"math"
	"reflect"
)

type Photon struct {
	name             string           // inherit
	timeline         *kernel.Timeline // inherit
	wavelength       float64
	location         *QuantumChannel        //tmp
	encodingType     map[string]interface{} // temp
	quantumState     []complex128
	entangledPhotons []*Photon //future []*Photon
}

func (photon *Photon) _init() {
	if photon.encodingType == nil {
		photon.encodingType = polarization()
	}
	if photon.quantumState == nil {
		photon.quantumState = []complex128{complex128(1), complex128(0)}
	}
	photon.entangledPhotons = []*Photon{photon}
}

func (photon *Photon) entangle(photon2 *Photon) {
	photon.entangledPhotons = append(photon.entangledPhotons, photon2)
	// need to do in entangle experience
}

func (photon *Photon) randomNoise(noise float64) {
	angle := noise * 2 * math.Pi
	photon.quantumState = []complex128{complex(math.Cos(angle), 0), complex(math.Sin(angle), 0)}
}

func (photon *Photon) setState(state []complex128) {
	for _, entangle := range photon.entangledPhotons {
		entangle.quantumState = state
	}
}

func (photon *Photon) measure(basis *Basis, prob float64) int {
	// only work for BB84
	state := oneToTwo(photon.quantumState) // 1-D array to 2-D array
	u := (*basis)[0]
	v := (*basis)[1]
	// measurement operator
	M0 := outer(arrayConj(u), u)
	M1 := outer(arrayConj(v), v)
	//projector0 := Basis{{1}}
	//projector1 := M1
	var projector0 *Basis
	var projector1 *Basis
	for _, p := range photon.entangledPhotons {
		if reflect.DeepEqual(p, photon) {
			projector0 = kron(&Basis{{1}}, M0)
			projector1 = kron(&Basis{{1}}, M1)
		} else {
			projector0 = kron(&Basis{{1}}, &Basis{{1, 0}, {0, 1}})
			projector1 = kron(&Basis{{1}}, &Basis{{1, 0}, {0, 1}})
		}
	}

	tmp := matMul(state.conj().transpose(), projector0.conj().transpose())
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
	for _, p := range photon.entangledPhotons {
		p.quantumState = newState
	}
	return result
}
