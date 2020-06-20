package quantum

import (
	"kernel"
)

type OpticalChannel struct {
	name        string           // inherit
	timeline    *kernel.Timeline // inherit
	attenuation float64
	//temperature          float64
	polarizationFidelity float64
	lightSpeed           float64 // (measured in m/ps)
	//chromaticDispersion  float64 // tmp
	distance float64
}
