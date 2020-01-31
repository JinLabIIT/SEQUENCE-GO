package quantum

import (
	"github.com/leesper/go_rng"
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
	"math/cmplx"
	"math/rand"
	"reflect"
	"strconv"
)

//node-> state int: 0 or 1
type Basis [][]complex128

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

func (photon *Photon) randomNoise() {
	angle := rand.Float64() * 2 * math.Pi
	photon.quantumState = []complex128{complex(math.Cos(angle), 0), complex(math.Sin(angle), 0)}
}

func (photon *Photon) setState(state []complex128) {
	for _, entangle := range photon.entangledPhotons {
		entangle.quantumState = state
	}
}

func (photon *Photon) measure(basis *Basis) int {
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
	if rand.Float64() > prob0 {
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

// quantumchannel functions
type QuantumChannel struct {
	OpticalChannel
	name          string           // inherit
	timeline      *kernel.Timeline // inherit
	sender        *LightSource
	receiver      *QSDetector //tmp
	depoCount     int
	photonCounter int
}

func (qc *QuantumChannel) setSender(sender *LightSource) {
	qc.sender = sender
}

func (qc *QuantumChannel) setReceiver(receiver *QSDetector) {
	qc.receiver = receiver
}

func (qc *QuantumChannel) get(photon *Photon) {
	loss := qc.distance * qc.attenuation
	chancePhotonKept := math.Pow(10, loss/-10)
	// check if photon kept
	if rand.Float64() < chancePhotonKept { // numpy.random.random_sample()
		qc.photonCounter += 1
		if rand.Float64() > qc.polarizationFidelity && photon.encodingType["name"] == "polarization" {
			photon.randomNoise()
			qc.depoCount += 1
		}
		futureTime := qc.timeline.Now() + uint64(math.Round(qc.distance/qc.lightSpeed))
		message := kernel.Message{"photon": photon}
		process := kernel.Process{Fnptr: qc.receiver.get, Message: message, Owner: qc.timeline}
		event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
		qc.timeline.Schedule(&event)
	}
}

// classical channel
type ClassicalChannel struct {
	OpticalChannel
	name     string           // inherit
	timeline *kernel.Timeline // inherit
	ends     []*Node          // ends must equal to 2
	delay    float64
}

func (cc *ClassicalChannel) _init() {
	if cc.delay == 0 {
		cc.delay = cc.distance / cc.lightSpeed
	}
}

func (cc *ClassicalChannel) addEnd(node *Node) {
	if exists(cc.ends, node) {
		panic("already have endpoint " + node.name)
	}
	if len(cc.ends) == 2 {
		panic("channel already has 2 endpoints")
	}
	cc.ends = append(cc.ends, node)
}

func (cc *ClassicalChannel) transmit(msg string, source *Node) {
	if !exists(cc.ends, source) {
		panic("no endpoint " + source.name)
	}
	/*	var receiver *Node
		for _, e := range cc.ends { // ?
			if e != source {
				receiver = e
			}
		}*/
	message := kernel.Message{"message": msg}
	futureTime := cc.timeline.Now() + uint64(math.Round(cc.delay))
	process := kernel.Process{Fnptr: source.receiveMessage, Message: message, Owner: cc.timeline}
	event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
	cc.timeline.Schedule(&event)
}

// lightSource function
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

type QSDetector struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	encodingType   map[string]interface{}
	detectors      []*Detector // tmp
	splitter       *BeamSplitter
	_switch        *Switch
	interferometer *Interferometer
}

func (qsd *QSDetector) _init() {
	if qsd.encodingType == nil {
		qsd.encodingType = polarization()
	}
	//fmt.Println(qsd.encodingType["name"])
	if (qsd.encodingType["name"] == "polarization" && len(qsd.detectors) != 2) ||
		(qsd.encodingType["name"] == "timeBin" && len(qsd.detectors) != 3) {
		panic("invalid number of detectors specified")
	}
	for i := range qsd.detectors {
		if !reflect.DeepEqual(qsd.detectors[i], Detector{}) { // question mark
			qsd.detectors[i].timeline = qsd.timeline
		} else {
			qsd.detectors[i] = &Detector{}
		}
	}
	if qsd.encodingType["name"] == "polarization" {
		bs := BeamSplitter{timeline: qsd.timeline}
		bs._init()
		qsd.splitter = &bs
	} else if qsd.encodingType["name"] == "timeBin" {
		qsd.interferometer = &Interferometer{timeline: qsd.timeline}
		qsd.interferometer.detectors = qsd.detectors[1:]
		qsd._switch = &Switch{timeline: qsd.timeline}
		qsd._switch.receiver = make([]interface{}, 0)
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.detectors[0])
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.interferometer)
		qsd._switch.typeList = []int{1, 0}
	} else {
		panic("invalid encoding type for QSDetector " + qsd.name)
	}
}

func (qsd *QSDetector) init() {
	for _, d := range qsd.detectors {
		if !reflect.DeepEqual(d, Detector{}) {
			d.init()
		}
	}
}

func (qsd *QSDetector) get(message kernel.Message) {
	photon := message["photon"].(*Photon)
	if qsd.encodingType["name"].(string) == "polarization" {
		detector := qsd.splitter.get(photon)
		//if detector == 0 || detector == 1 {
		//detector = qsd.splitter.get(photon)//test
		qsd.detectors[detector].get(kernel.Message{"darkGet": false}) //??????
		//}
	} else if qsd.encodingType["name"].(string) == "timeBin" {
		qsd._switch.get(photon)
	}
}

func (qsd *QSDetector) clearDetectors(message kernel.Message) {
	for _, d := range qsd.detectors {
		d.photonTimes = []uint64{}
	}
}

func (qsd *QSDetector) getPhotonTimes() [][]uint64 {
	var times [][]uint64
	for _, d := range qsd.detectors {
		if !reflect.DeepEqual(d, Detector{}) {
			times = append(times, d.photonTimes)
		} else {
			times = append(times, []uint64{})
		}
	}
	return times
}

func (qsd *QSDetector) setBasis(basis *Basis) {
	qsd.splitter.setBasis(basis)
}

func (qsd *QSDetector) turnOffDetectors() {
	for _, d := range qsd.detectors {
		d.on = false
	}
}

func (qsd *QSDetector) turnOnDetectors() {
	for _, d := range qsd.detectors {
		if !(d.on) {
			d.init()
			d.on = true
		}
	}
}

type Detector struct {
	name              string           //inherit
	timeline          *kernel.Timeline //inherit
	efficiency        float64
	darkCount         float64
	countRate         float64
	timeResolution    uint64
	photonTimes       []uint64
	nextDetectionTime uint64
	photonCounter     int
	on                bool
}

func (d *Detector) _init() {
	if d.efficiency == 0 {
		d.efficiency = 1
	}
	if d.countRate == 0 {
		d.countRate = math.MaxFloat64 // measured in Hz
	}
	if d.timeResolution == 0 {
		d.timeResolution = 1
	}
	d.on = true
}

func (d *Detector) init() {
	d.addDarkCount(kernel.Message{})
}

func (d *Detector) get(message kernel.Message) {
	darkGet := message["darkGet"].(bool)
	d.photonCounter += 1
	now := d.timeline.Now()
	if (rand.Float64() < d.efficiency || darkGet) || (now > d.nextDetectionTime) {
		time := (now / d.timeResolution) * d.timeResolution
		d.photonTimes = append(d.photonTimes, time)
		d.nextDetectionTime = now + uint64(math.Pow10(12)/d.countRate)
	}
}

func (d *Detector) addDarkCount(message kernel.Message) {
	if d.on {
		timeToNext := uint64(rand.ExpFloat64()/d.darkCount) * uint64(math.Pow10(12))
		time := timeToNext + d.timeline.Now()
		message1 := kernel.Message{}
		process1 := kernel.Process{Fnptr: d.addDarkCount, Message: message1, Owner: d.timeline}
		event1 := kernel.Event{Time: time, Process: &process1, Priority: 0}
		message2 := kernel.Message{"darkGet": true}
		process2 := kernel.Process{Fnptr: d.get, Message: message2, Owner: d.timeline}
		event2 := kernel.Event{Time: time, Process: &process2, Priority: 0}
		d.timeline.Schedule(&event1)
		d.timeline.Schedule(&event2)
	}
}

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

type Interferometer struct {
	timeline       *kernel.Timeline // inherit
	pathDifference int
	phaseError     float64
	detectors      []*Detector
}

func (inf *Interferometer) get(photon *Photon) {
	detectorNum := rand.Intn(2)
	quantumState := photon.quantumState
	time := 0
	random := rand.Float64()

	if quantumState[0] == complex(1, 0) && quantumState[1] == complex(0, 0) { // early
		if random <= 0.5 {
			time = 0
		} else {
			time = inf.pathDifference
		}
	}
	if quantumState[0] == complex(0, 0) && quantumState[1] == complex(1, 0) { // late
		if random <= 0.5 {
			time = inf.pathDifference
		} else {
			time = 2 * inf.pathDifference
		}
	}
	if rand.Float64() < inf.phaseError {
		quantumState = multiply([]float64{1, -1}, quantumState) // list??
	}
	if quantumState[0] == complex(math.Sqrt(0.5), 0) && quantumState[1] == complex(math.Sqrt(0.5), 0) { // early + late
		if random <= 0.25 {
			time = 0
		} else if random <= 0.5 {
			time = 2 * inf.pathDifference
		} else if detectorNum == 0 {
			time = inf.pathDifference
		} else {
			return
		}
		if quantumState[0] == complex(math.Sqrt(0.5), 0) && quantumState[1] == complex(math.Sqrt(-0.5), 0) { // early - late
			if random <= 0.25 {
				time = 0
			} else if random <= 0.5 {
				time = 2 * inf.pathDifference
			} else if detectorNum == 1 {
				time = inf.pathDifference
			} else {
				return
			}
		}
		message := kernel.Message{}
		process := kernel.Process{Fnptr: inf.detectors[detectorNum].get, Message: message, Owner: inf.timeline}
		event := kernel.Event{Time: inf.timeline.Now() + uint64(time), Process: &process, Priority: 0}
		inf.timeline.Schedule(&event)
	}
}

type Switch struct {
	timeline  *kernel.Timeline
	receiver  []interface{} // Interferometer Detector
	startTime uint64
	frequency float64
	stateList []int // tmp
	typeList  []int //0: Interferometer 1: Detector ???
}

func (_switch *Switch) addReceiver(entity *interface{}) {
	_switch.receiver = append(_switch.receiver, entity)
}

func (_switch *Switch) setState(state int) {
	_switch.stateList = []int{state}
}

func (_switch *Switch) get(photon *Photon) {
	index := int(float64(_switch.timeline.Now()-_switch.startTime) * _switch.frequency * math.Pow10(-12))
	if index < 0 || index >= len(_switch.stateList) {
		index = 0
	}
	receiver := _switch.receiver[_switch.stateList[index]]
	// check if receiver is detector, if we're using time bin, and if the photon is "late" to schedule measurement
	if _switch.typeList[index] == 1 { //???
		if photon.encodingType["name"] == "timeBin" && photon.measure(photon.encodingType["bases"].([]*Basis)[0]) == 1 {
			time := _switch.timeline.Now() + photon.encodingType["binSeparation"].(uint64)
			message := kernel.Message{}
			process := kernel.Process{Fnptr: receiver.(*Detector).get, Message: message, Owner: _switch.timeline}
			event := kernel.Event{Time: time, Priority: 0, Process: &process}
			_switch.timeline.Schedule(&event)
		} else {
			receiver.(*Detector).get(kernel.Message{})
		}
	} else {
		receiver.(*Interferometer).get(photon)
	}

}

type Node struct {
	name       string           // inherit
	timeline   *kernel.Timeline // inherit
	components map[string]interface{}
	count      []int
	message    kernel.Message //temporary storage for message received through classical channel
	protocol   interface{}    //
	splitter   *BeamSplitter
	receiver   *Node
}

func (node *Node) sendQubits(basisList, bitList []int, sourceName string) {
	encodingType := node.components[sourceName].(*LightSource).encodingType //???
	stateList := make(Basis, 0, len(bitList))
	for i, bit := range bitList {
		basis := (encodingType["bases"].([][][]complex128))[basisList[i]]
		state := basis[bit]
		stateList = append(stateList, state)
	}
	node.components[sourceName].(*LightSource).emit(&stateList)
}

func (node *Node) getBits(lightTime float64, startTime uint64, frequency float64, detectorName string) []int {
	encodingType := node.components[detectorName].(*QSDetector).encodingType
	length := int(math.Round(lightTime * frequency))
	bits := makeArray(length, -1) // -1 used for invalid bits
	if encodingType["name"] == "polarization" {
		// determine indices from detection times and record bits
		detectionTimes := node.components[detectorName].(*QSDetector).getPhotonTimes()
		for _, time := range detectionTimes[0] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				bits[index] = 0
			}
		}
		for _, time := range detectionTimes[1] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				if bits[index] == 0 {
					bits[index] = -1
				} else {
					bits[index] = 1
				}
			}
		}
		// need to be deleted
		count0 := 0
		count1 := 1
		for _, i := range bits {
			if i == 0 {
				count0++
			}
			if i == 1 {
				count1++
			}
		}
		cc := count0 + count1
		cc = cc
		// need to be deleted
		return bits
	} else if encodingType["name"] == "timeBin" {
		detectionTimes := node.components[detectorName].(*QSDetector).getPhotonTimes()
		binSeparation, _ := encodingType["binSeparation"].(float64)
		// single detector (for early, late basis) times
		for _, time := range detectionTimes[0] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				if math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
					bits[index] = 0
				} else if math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime)-(float64(time)-binSeparation)) < binSeparation/2 {
					bits[index] = 1
				}
			}
		}
		for _, time := range detectionTimes[1] {
			time -= uint64(binSeparation)
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if (0 <= index && index < len(bits)) && math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
				if bits[index] == -1 {
					bits[index] = 0
				} else {
					bits[index] = -1
				}
			}
		}
		for _, time := range detectionTimes[2] {
			time -= uint64(binSeparation)
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if (0 <= index && index < len(bits)) && math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
				if bits[index] == -1 {
					bits[index] = 1
				} else {
					bits[index] = -1
				}
			}
		}
		return bits
	} else {
		panic("Invalid encoding type for node " + node.name)
	}
}

func (node *Node) setBases(basisList []int, startTime uint64, frequency float64, detectorName string) {
	encodingType := node.components[detectorName].(*QSDetector).encodingType
	basisStartTime := startTime - uint64(math.Pow10(12)/(2*frequency))
	if encodingType["name"] == "polarization" {
		splitter := node.components[detectorName].(*QSDetector).splitter
		splitter.startTime = basisStartTime
		splitter.frequency = frequency

		splitterBasisList := make([]*Basis, 0, len(basisList))
		for _, d := range basisList {
			base := encodingType["bases"].([][][]complex128)
			tmp := Basis{base[d][0], base[d][1]}
			splitterBasisList = append(splitterBasisList, &tmp)
		}
		splitter.basisList = splitterBasisList
	} else if encodingType["name"] == "timeBin" {
		_switch := node.components[detectorName].(*QSDetector)._switch
		_switch.startTime = basisStartTime
		_switch.frequency = frequency
		_switch.stateList = basisList
	} else {
		panic("Invalid encoding type for node " + node.name)
	}
}

func (node *Node) getSourceCount() interface{} { // tmp
	source := node.components["lightSource"]
	return source
}

func (node *Node) sendMessage(msg string, channel string) {
	fmt.Println("sendMessage " + strconv.FormatUint(node.timeline.Now(), 10))
	node.components[channel].(*ClassicalChannel).transmit(msg, node.receiver)
}

func (node *Node) receiveMessage(message kernel.Message) {
	node.message = message
	node.protocol.(*BB84).receivedMessage()
}

// help functions
func exists(slice []*Node, val *Node) bool {
	for _, item := range slice {
		if item == val { // question mark
			return true
		}
	}
	return false
}

func multiply(base []float64, state []complex128) []complex128 { // 2*2 matrix * 2*2 matrix
	a := complex(base[0], 0) * state[0]
	b := complex(base[1], 0) * state[1]
	return []complex128{a, b}
}

func makeArray(length int, value int) []int {
	results := make([]int, length)
	for i := 0; i < length; i++ {
		results[i] = value
	}
	return results
}

func outer(a, b []complex128) *Basis { // assume a and b are m*1 and 1*n matrix
	result := make(Basis, len(a))
	for i, c := range a {
		for _, d := range b {
			result[i] = append(result[i], c*d)
		}
	}
	return &result
}

func kron(a, b *Basis) *Basis { // a->m*n b->i*j
	rowA := len(*a)
	rowB := len(*b)
	colA := len((*a)[0])
	colB := len((*b)[0])
	result := make(Basis, rowA*rowB)
	for m := 0; m < rowA; m++ {
		for n := 0; n < colA; n++ {
			for i := 0; i < rowB; i++ {
				for j := 0; j < colB; j++ {
					result[m*rowB+i] = append(result[m*rowB+i], (*a)[m][n]*(*b)[i][j])
				}
			}
		}
	}
	return &result
}

func (basis *Basis) transpose() *Basis {
	m := len(*basis)
	n := len((*basis)[0])
	result := make(Basis, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			result[j] = append(result[j], (*basis)[i][j])
		}
	}
	return &result
}

func (basis *Basis) conj() *Basis {
	result := make(Basis, len(*basis))
	for i := 0; i < len(*basis); i++ {
		for j := 0; j < len((*basis)[0]); j++ {
			result[i] = append(result[i], cmplx.Conj((*basis)[i][j]))
		}
	}
	return &result
}

func matMul(a, b *Basis) *Basis { // Matrix multiplication a->m*n b->n*p
	m := len(*a)
	n := len((*a)[0])
	p := len((*b)[0])
	if n != len(*b) {
		panic("the columns of first matrix must equal to the rows of the second matrix")
	}
	result := make(Basis, m)
	for i := 0; i < m; i++ {
		for j := 0; j < p; j++ {
			val := helpMatMul(a, b, i, j) //a[i][]*b[][j]
			result[i] = append(result[i], val)
		}
	}
	return &result
}

func helpMatMul(a, b *Basis, aIndex int, bIndex int) complex128 { // a[i][] * b[][j]
	var result complex128
	for i := 0; i < len((*a)[0]); i++ {
		result += (*a)[aIndex][i] * (*b)[i][bIndex]
	}
	return result
}

func oneToTwo(a []complex128) *Basis { //one dimension to two dimension array
	result := make(Basis, 2)
	for i := 0; i < len(a); i++ {
		result[i] = []complex128{a[i]}
	}
	return &result
}

func divide(a *Basis, b float64) []complex128 { //1*n matrix divided by float
	if b == 0 {
		panic("can not divided by ZERO")
	}
	result := make([]complex128, 2)
	result[0] = (*a)[0][0] / complex(b, 0)
	result[1] = (*a)[1][0] / complex(b, 0)
	return result
}

func arrayConj(arr []complex128) []complex128 {
	for i, ele := range arr {
		arr[i] = cmplx.Conj(ele)
	}
	return arr
}
