package quantum

import (
	rng "github.com/leesper/go_rng"
	"kernel"
	"math"
)

type BeamSplitter struct {
	timeline  *kernel.Timeline // inherit
	basis     [2][2]complex128
	fidelity  float64
	startTime uint64
	frequency float64
	basisList [][2][2]complex128
	rng       *rng.UniformGenerator
}

func (bs *BeamSplitter) _init() {
	bs.rng = rng.NewUniformGenerator(123)
	//if bs.basis == nil {
	bs.basis = [2][2]complex128{{complex128(1), complex128(0)}, {complex128(0), complex128(1)}}
	//}
	if bs.basisList == nil {
		bs.basisList = [][2][2]complex128{bs.basis}
	}
	if bs.fidelity == 0 {
		bs.fidelity = 1
	}
}

func (bs *BeamSplitter) get(photon *Photon) int {
	if bs.rng.Float64() < bs.fidelity {
		index := int(float64(bs.timeline.Now()-bs.startTime) * bs.frequency * math.Pow10(-12))
		prob := bs.rng.Float64()
		if 0 <= index && index < len(bs.basisList) {
			return photon.measure(&bs.basisList[index], prob)
		} else {
			return photon.measure(&bs.basisList[0], prob)
		}
	} else {
		return -1
	}
}

func (bs *BeamSplitter) setBasis(basis *[2][2]complex128) {
	bs.basisList = [][2][2]complex128{*basis}
}
