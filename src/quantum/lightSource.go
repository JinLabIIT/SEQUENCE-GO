package quantum

import (
	rng "github.com/leesper/go_rng"
	"kernel"
	"math"
	"math/rand"
)

type LightSource struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	frequency      float64
	wavelength     float64
	lineWidth      float64
	meanPhotonNum  float64
	encodingType   map[string]interface{}
	directReceiver *QuantumChannel
	phaseError     float64
	photonCounter  int
	poisson        *rng.PoissonGenerator
}

func (ls *LightSource) _init() {
	if ls.wavelength == 0 {
		ls.wavelength = 1550
	}
	if ls.encodingType == nil {
		ls.encodingType = polarization()
	}
}

// can be optimized later
func (ls *LightSource) emit(stateList *Basis) {
	//fmt.Println("emit message")
	time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i, state := range *stateList {
		numPhotons := ls.poisson.Poisson(ls.meanPhotonNum) //question mark
		if numPhotons > 0 {
			if rand.Float64() < ls.phaseError {
				multiply([]float64{1.0, -1.0}, state)
			}
			message := kernel.Message{"stateList": stateList, "numPhotons": numPhotons, "state": state, "index": i + 1}
			process := kernel.Process{Fnptr: ls._emit, Message: message, Owner: ls.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			ls.timeline.Schedule(&event)
			break
		}
		time += sep
	}
}

func (ls *LightSource) _emit(message kernel.Message) {
	//fmt.Println("_emit")
	stateList := message["stateList"].(*Basis)
	numPhotons := message["numPhotons"].(int64)
	state := message["state"].([]complex128)
	index := message["index"].(int)
	time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i := 0; i < int(numPhotons); i++ {
		wavelength := ls.lineWidth*rand.NormFloat64() + ls.wavelength
		newPhoton := Photon{timeline: ls.timeline, wavelength: wavelength, location: ls.directReceiver, encodingType: ls.encodingType, quantumState: state}
		newPhoton._init()
		ls.directReceiver.get(&newPhoton)
		ls.photonCounter += 1
	}
	time += sep
	for index < len(*stateList) {
		numPhotons := ls.poisson.Poisson(ls.meanPhotonNum)
		if numPhotons > 0 {
			state = (*stateList)[index]
			if rand.Float64() < ls.phaseError {
				multiply([]float64{1.0, -1.0}, state)
			}
			message := kernel.Message{"stateList": stateList, "numPhotons": numPhotons, "state": state, "index": index + 1}
			process := kernel.Process{Fnptr: ls._emit, Message: message, Owner: ls.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			ls.timeline.Schedule(&event)
			break
		}
		index += 1
		time += sep
	}
}

func (ls *LightSource) assignReceiver(receiver *QuantumChannel) {
	ls.directReceiver = receiver
}
