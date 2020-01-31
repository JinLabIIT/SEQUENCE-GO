package quantum

import (
	"kernel"
	"math"
	"math/rand"
)

type BeamSplitter struct {
	timeline  *kernel.Timeline // inherit
	basis     *Basis
	fidelity  float64
	startTime uint64
	frequency float64
	basisList []*Basis
}

func (bs *BeamSplitter) _init() {
	if bs.basis == nil {
		bs.basis = &Basis{{complex128(1), complex128(0)}, {complex128(0), complex128(1)}}
	}
	if bs.basisList == nil {
		bs.basisList = []*Basis{bs.basis}
	}
	if bs.fidelity == 0 {
		bs.fidelity = 1
	}
}

func (bs *BeamSplitter) get(photon *Photon) int {
	if rand.Float64() < bs.fidelity {
		index := int(float64(bs.timeline.Now()-bs.startTime) * bs.frequency * math.Pow10(-12))
		if 0 <= index && index < len(bs.basisList) {
			return photon.measure(bs.basisList[index])
		} else {
			return photon.measure(bs.basisList[0])
		}
	} else {
		return -1
	}
}

func (bs *BeamSplitter) setBasis(basis *Basis) {
	bs.basisList = []*Basis{basis}
}
