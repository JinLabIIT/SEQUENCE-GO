package quantum

import (
	"kernel"
	"math"
)

type OpticalChannel struct {
	name                 string           // inherit
	timeline             *kernel.Timeline // inherit
	attenuation          float64
	temperature          float64
	polarizationFidelity float64
	lightSpeed           float64
	chromaticDispersion  float64 // tmp
	distance             float64
}

func (oc *OpticalChannel) _init() {
	if oc.polarizationFidelity == 0 {
		oc.polarizationFidelity = 1
	}
	if oc.lightSpeed == 0 {
		oc.lightSpeed = 3 * math.Pow10(-4) // used for photon timing calculations (measured in m/ps)
	}
	if oc.chromaticDispersion == 0 {
		oc.chromaticDispersion = 17 // measured in ps
	}
}
