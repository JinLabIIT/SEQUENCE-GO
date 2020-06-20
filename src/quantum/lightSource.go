package quantum

import (
	rng "github.com/leesper/go_rng"
	"kernel"
	"math"
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
	rng            *rng.UniformGenerator
	grng           *rng.GaussianGenerator
}

func (ls *LightSource) init() {
	ls.rng = rng.NewUniformGenerator(123)
	ls.grng = rng.NewGaussianGenerator(123)
	ls.poisson = rng.NewPoissonGenerator(123)
}

// can be optimized later
func (ls *LightSource) emit(stateList *Basis) {
	//fmt.Println("emit message")
	time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i, state := range *stateList {
		numPhotons := ls.poisson.Poisson(ls.meanPhotonNum) //question mark
		if numPhotons > 0 {
			if ls.rng.Float64() < ls.phaseError {
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
	Time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i := 0; i < int(numPhotons); i++ {
		//wavelength := ls.lineWidth*ls.grng.Gaussian(0, 1) + ls.wavelength
		//creatPhoton = time.Now().UnixNano()
		//newPhoton := Photon{location: ls.directReceiver, encodingType: ls.encodingType, quantumState: state}
		newPhoton := Photon{encodingType: ls.encodingType, quantumState: state}
		//endCreate = time.Now().UnixNano()
		newPhoton._init()
		//schedTime = time.Now().UnixNano()
		ls.directReceiver.get(&newPhoton)
		//schedEndTime = time.Now().UnixNano()
		ls.photonCounter += 1
	}
	Time += sep
	counter := 0
	for index < len(*stateList) {
		numPhotons := ls.poisson.Poisson(ls.meanPhotonNum)
		if numPhotons > 0 {
			state = (*stateList)[index]
			if ls.rng.Float64() < ls.phaseError {
				multiply([]float64{1.0, -1.0}, state)
			}
			message := kernel.Message{"stateList": stateList, "numPhotons": numPhotons, "state": state, "index": index + 1}
			process := kernel.Process{Fnptr: ls._emit, Message: message, Owner: ls.timeline}
			event := kernel.Event{Time: Time, Process: &process, Priority: 0}
			//secSchedS = time.Now().UnixNano()
			ls.timeline.Schedule(&event)
			//secSchedE = time.Now().UnixNano()
			break
		}
		counter++
		index += 1
		Time += sep
	}
}

func (ls *LightSource) assignReceiver(receiver *QuantumChannel) {
	ls.directReceiver = receiver
}
