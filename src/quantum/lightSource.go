package quantum

import (
	rng "github.com/leesper/go_rng"
	"kernel"
	"math"
)

type LightSource struct {
	name            string           // inherit
	timeline        *kernel.Timeline // inherit
	frequency       float64
	wavelength      float64
	lineWidth       float64
	meanPhotonNum   float64
	encodingType    map[string]interface{}
	directReceiver  *QuantumChannel
	phaseError      float64
	photonCounter   int
	poisson         *rng.PoissonGenerator
	rng             *rng.UniformGenerator
	grng            *rng.GaussianGenerator
	event_counter   int
	first_emit_time uint64
}

func (ls *LightSource) init(seed int64) {
	ls.rng = rng.NewUniformGenerator(seed)
	ls.grng = rng.NewGaussianGenerator(seed)
	ls.poisson = rng.NewPoissonGenerator(seed)
}

// can be optimized later
func (ls *LightSource) emit(stateList *[][]complex128) {
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
	if ls.first_emit_time == 0 {
		//ls.first_emit_time = ls.timeline.Now()
	}
	ls.event_counter += 1
	//if ls.timeline.Now() >= ls.timeline.OutputTime {
	//	fmt.Println(ls.timeline.Now(), ls.event_counter)
	//}
	stateList := message["stateList"].(*[][]complex128)
	numPhotons := message["numPhotons"].(int64)
	state := message["state"].([]complex128)
	index := message["index"].(int)
	Time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i := 0; i < int(numPhotons); i++ {
		//wavelength := ls.lineWidth*ls.grng.Gaussian(0, 1) + ls.wavelength
		//newPhoton := Photon{location: ls.directReceiver, encodingType: ls.encodingType, quantumState: state}
		newPhoton := Photon{encodingType: ls.encodingType, quantumState: [2]complex128{state[0], state[1]}}
		newPhoton._init()
		ls.directReceiver.get(&newPhoton)
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
			ls.timeline.Schedule(&event)
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
