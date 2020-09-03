package quantum

import (
	rng "github.com/leesper/go_rng"
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
	"math/rand"
	"sync"
)

type ls struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	frequency      float64
	wavelength     float64
	lineWidth      float64
	meanPhotonNum  float64
	encodingType   map[string]interface{}
	directReceiver *qc
	phaseError     float64
	photonCounter  int
	poisson        *rng.PoissonGenerator
	bs             *[][]complex128
	tran           int
	lookahead      uint64
}

func (l *ls) _init() {
	if l.wavelength == 0 {
		l.wavelength = 1550
	}
	if l.encodingType == nil {
		l.encodingType = polarization()
	}
}

func (l *ls) transmit(message kernel.Message) {
	l.tran++
	time := l.timeline.Now()
	msg := kernel.Message{"stateList": l.bs}
	process1 := kernel.Process{Fnptr: l.emit, Message: msg, Owner: l.timeline}
	event1 := kernel.Event{Time: time + uint64(math.Pow10(2)), Process: &process1, Priority: 0}
	l.timeline.Schedule(&event1)

	process := kernel.Process{Fnptr: l.transmit, Message: kernel.Message{}, Owner: l.timeline}
	event := kernel.Event{Time: time + uint64(math.Pow10(2)), Process: &process, Priority: 0}
	l.timeline.Schedule(&event)
}

func (l *ls) emit(message kernel.Message) {
	//fmt.Println("emit message")
	time := l.timeline.Now()
	stateList := message["stateList"].(*[][]complex128)
	//sep := uint64(math.Round(math.Pow10(12) / l.frequency))
	sep := uint64(0)
	for i, state := range *stateList {
		numPhotons := l.poisson.Poisson(l.meanPhotonNum) //question mark
		if numPhotons > 0 {
			if rand.Float64() < l.phaseError {
				multiply([]float64{1.0, -1.0}, state)
			}
			message := kernel.Message{"stateList": stateList, "numPhotons": numPhotons, "state": state, "index": i + 1}
			process := kernel.Process{Fnptr: l._emit, Message: message, Owner: l.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			l.timeline.Schedule(&event)
			break
		}
		time += sep
	}
}

func (l *ls) _emit(message kernel.Message) {
	//fmt.Println("_emit")
	stateList := message["stateList"].(*[][]complex128)
	numPhotons := message["numPhotons"].(int64)
	state := message["state"].([]complex128)
	index := message["index"].(int)
	time := l.timeline.Now()
	//sep := uint64(math.Round(math.Pow10(12) / l.frequency))
	sep := uint64(0)
	for i := 0; i < int(numPhotons); i++ {
		//wavelength := l.lineWidth*rand.NormFloat64() + l.wavelength
		//newPhoton := Photon{timeline: l.timeline, wavelength: wavelength, location: l.directReceiver, encodingType: l.encodingType, quantumState: state}
		//newPhoton._init()
		newPhoton := rand.Float64()
		message := kernel.Message{"photon": newPhoton}
		process := kernel.Process{Fnptr: l.directReceiver.get, Message: message, Owner: l.directReceiver.timeline}
		event := kernel.Event{Time: time + l.lookahead, Process: &process, Priority: 0}
		l.timeline.Schedule(&event)
		l.photonCounter += 1
	}
	time += sep
	for index < len(*stateList) {
		numPhotons := l.poisson.Poisson(l.meanPhotonNum)
		if numPhotons > 0 {
			state = (*stateList)[index]
			if rand.Float64() < l.phaseError {
				multiply([]float64{1.0, -1.0}, state)
			}
			message := kernel.Message{"stateList": stateList, "numPhotons": numPhotons, "state": state, "index": index + 1}
			process := kernel.Process{Fnptr: l._emit, Message: message, Owner: l.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			l.timeline.Schedule(&event)
			break
		}
		index += 1
		time += sep
	}
}

type qc struct {
	OpticalChannel
	name          string           // inherit
	timeline      *kernel.Timeline // inherit
	sender        *LightSource
	receiver      *qsd //tmp
	depoCount     int
	photonCounter int
}

func (q *qc) get(message kernel.Message) {
	photon := message["photon"].(float64)
	//loss := q.distance * q.attenuation
	//chancePhotonKept := math.Pow(10, loss/-10)
	// check if photon kept
	//if rand.Float64() < chancePhotonKept { // numpy.random.random_sample()
	q.photonCounter += 1

	futureTime := q.timeline.Now() + uint64(math.Round(q.distance/q.lightSpeed))
	msg := kernel.Message{"photon": photon}
	process := kernel.Process{Fnptr: q.receiver.get, Message: msg, Owner: q.timeline}
	event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
	q.timeline.Schedule(&event)
	//}
}

type qsd struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	encodingType   map[string]interface{}
	detectors      []*Detector // tmp
	splitter       *BeamSplitter
	_switch        *Switch
	interferometer *Interferometer
}

func (qsd *qsd) get(message kernel.Message) {
	photon := message["photon"].(float64)

	if photon > 0.5 {
		//fmt.Println("1")
	} else {
		//fmt.Println("0")
	}
}

/*func MakeStateList() *Basis {
	numPulses := 10000
	basisList := choice([]int{0, 1}, numPulses)
	bitList := choice([]int{0, 1}, numPulses)
	encodingType := polarization()
	stateList := make(Basis, numPulses)
	for i, bit := range bitList {
		basis := (encodingType["bases"].([][][]complex128))[basisList[i]]
		state := basis[bit]
		stateList[i] = state
	}
	return &stateList
}*/

func addFunction(n int, wg *sync.WaitGroup) {
	result := 0
	for i := 0; i < n; i++ {
		result = result + 1
	}
	wg.Done()
}

func createRandom(n int, wg *sync.WaitGroup) {
	uniform := rng.NewExpGenerator(1)
	for i := 0; i < n; i++ {
		uniform.Exp(1)
	}
	fmt.Println(n)
	wg.Done()
}
