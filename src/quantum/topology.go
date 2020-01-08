package quantum

import (
	"github.com/leesper/go_rng"
	"kernel"
	"math"
	"math/cmplx"
	"math/rand"
	"os"
	"reflect"
)

type Dense struct {
	basis [][]complex128
}

//node-> state int: 0 or 1
type Basis [][]complex128

type Photon struct {
	name             string           // inherit
	timeline         *kernel.Timeline // inherit
	wavelength       float64
	location         *QuantumChannel        //tmp
	encodingType     map[string]interface{} // temp
	quantumState     []complex128
	entangledPhotons *Photon //future []*Photon
}

func (photon *Photon) entangle(photon2 *Photon) {
	photon.entangledPhotons = photon2
}

func (photon *Photon) randomNoise() {
	angle := rand.Float64() * 2 * math.Pi
	photon.quantumState = []complex128{complex(math.Cos(angle), 0), complex(math.Sin(angle), 0)}
}

func (photon *Photon) setState(state []complex128) {
	photon.quantumState = state
}

func (photon *Photon) measure(basis *Basis) int {
	// only work for BB84
	state := oneToTwo(photon.quantumState)
	u := (*basis)[0]
	v := (*basis)[1]
	// measurement operator
	M0 := outer(u, u, true)
	M1 := outer(v, v, true)
	projector0 := M0
	projector1 := M1
	tmp := matMul(transpose(state, true), transpose(projector0, true))
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
		newState = (*divide(matMul(projector0, state), math.Sqrt(1-prob0)))[0]
	} else {
		newState = (*divide(matMul(projector1, state), math.Sqrt(1-prob0)))[0]
	}
	photon.quantumState = newState
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
	qc.setSender(sender)
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

func (qsd *QSDetector) turnOnDetector() {
	for _, d := range qsd.detectors {
		if !d.on {
			d.init()
			d.on = true
		}
	}
}

// classical channel
type ClassicalChannel struct {
	OpticalChannel
	name     string           // inherit
	timeline *kernel.Timeline // inherit
	ends     []*Node          // tmp
	delay    float64
}

func (cc *ClassicalChannel) addEnd(node *Node) {
	if exists(cc.ends, node) {
		panic("-1")
	}
	if len(cc.ends) == 2 {
		os.Exit(-1)
	}
	cc.ends = append(cc.ends, node)
}

func (cc *ClassicalChannel) transmit(msg string, source *Node) {
	if exists(cc.ends, source) {
		os.Exit(-1)
	}
	var receiver *Node
	for _, e := range cc.ends {
		if e != source {
			receiver = e
		}
	}
	message := kernel.Message{"message": msg}
	futureTime := cc.timeline.Now() + uint64(math.Round(cc.delay))
	process := kernel.Process{Fnptr: receiver.receiveMessage, Message: message, Owner: cc.timeline}
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
	encodingType   map[string]interface{} // tmp
	directReceiver *QuantumChannel        // tmp
	phaseError     float64                // tmp
	photonCounter  int
	poisson        *rng.PoissonGenerator
}

// can be optimized later
func (ls *LightSource) emit(stateList *Basis) { // tmp []int
	time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i, state := range *stateList {
		numPhotons := ls.poisson.Poisson(1) //question mark
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
	stateList := message["stateList"].([][]complex128)
	numPhotons := message["numPhotons"].(int)
	state := message["state"].([]complex128)
	index := message["index"].(int)
	time := ls.timeline.Now()
	sep := uint64(math.Round(math.Pow10(12) / ls.frequency))
	for i := 0; i < numPhotons; i++ {
		wavelength := ls.lineWidth*rand.NormFloat64() + ls.wavelength
		newPhoton := Photon{timeline: ls.timeline, wavelength: wavelength, location: ls.directReceiver, encodingType: ls.encodingType, quantumState: state}
		ls.directReceiver.get(&newPhoton)
		ls.photonCounter += 1
		time += sep
		for index < len(stateList) {
			numPhotons := ls.poisson.Poisson(1)
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
			index += 1
			time += sep
		}
	}
}

func (ls *LightSource) assignReceiver(receiver *QuantumChannel) {
	ls.directReceiver = receiver
}

type QSDetector struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	encodingType   map[string]interface{}
	detectors      []Detector // tmp
	splitter       *BeamSplitter
	_switch        *Switch
	interferometer *Interferometer
}

func (qsd *QSDetector) _init() {
	if (qsd.encodingType["name"] == "polarization" && len(qsd.detectors) != 2) ||
		(qsd.encodingType["name"] == "timeBin" && len(qsd.detectors) != 3) {
		os.Exit(-1)
	}
	for i := range qsd.detectors {
		if !reflect.DeepEqual(qsd.detectors[i], Detector{}) { // question mark
			qsd.detectors[i].timeline = qsd.timeline
		} else {
			qsd.detectors[i] = Detector{}
		}
	}
	if qsd.encodingType["name"] == "polarization" {
		// need to do
		qsd.splitter = &BeamSplitter{timeline: qsd.timeline}
	} else if qsd.encodingType["name"] == "timeBin" {
		// need to do
		qsd.interferometer = &Interferometer{timeline: qsd.timeline}
		qsd.interferometer.detectors = qsd.detectors[1:]
		qsd._switch = &Switch{timeline: qsd.timeline}
		qsd._switch.receiver = make([]interface{}, 0)
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.detectors[0])
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.interferometer)
		qsd._switch.typeList = []int{1, 0}
	} else {
		os.Exit(-1)
	}
}

func (qsd *QSDetector) init() {
	for _, d := range qsd.detectors {
		if reflect.DeepEqual(d, Detector{}) {
			d.init()
		}
	}
}

func (qsd *QSDetector) get(message kernel.Message) {
	photon := message["photon"].(*Photon)
	if qsd.encodingType["name"] == "polarization" {
		detector := qsd.splitter.get(photon)
		if detector == 0 || detector == 1 {
			qsd.detectors[qsd.splitter.get(photon)].get(kernel.Message{"darkGet": false})
		}
	} else if qsd.encodingType["name"] == "timeBin" {
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
			d.init() // ??
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

func (d *Detector) init() {
	d.addDarkCount(kernel.Message{})
}

func (d *Detector) get(message kernel.Message) {
	darkGet := message["bool"].(bool)
	d.photonCounter += 1
	now := d.timeline.Now()
	if (rand.Float64() < d.efficiency || darkGet) || now > d.nextDetectionTime {
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
		message2 := kernel.Message{"bool": true}
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
	pathDifference int              // tmp
	phaseError     float64          // tmp
	detectors      []Detector       // tmp
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
			process := kernel.Process{Fnptr: receiver.(Detector).get, Message: message, Owner: _switch.timeline}
			event := kernel.Event{Time: time, Priority: 0, Process: &process}
			_switch.timeline.Schedule(&event)
		}
	} else {
		receiver.(Interferometer).get(photon)
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
}

func (node *Node) sendQubits(basisList, bitList []int, sourceName string) {

	encodingType := node.components[sourceName].(LightSource).encodingType //???
	stateList := make(Basis, len(bitList))
	for i, bit := range bitList {
		basis := (encodingType["bases"].([]*Basis))[basisList[i]]
		state := (*basis)[bit]
		stateList[len(stateList)-i-1] = state
	}
	node.components[sourceName].(LightSource).emit(&stateList)
}

func (node *Node) sendPhotons(state complex128, num int, sourceName string) { // no need in BB84, complete later
	//stateList := makeArray(state, num)
	//node.components[sourceName].(LightSource).emit(stateList)
}

func (node *Node) getBits(lightTime float64, startTime uint64, frequency float64, detectorName string) []int {
	encodingType := node.components[detectorName].(QSDetector).encodingType // ???
	length := int(math.Round(lightTime * frequency))                        // -1 used for invalid bits
	bits := makeArray(length, -1)
	detectionTimes := node.components[detectorName].(QSDetector).getPhotonTimes()
	if encodingType["name"] == "polarization" {
		// determine indices from detection times and record bits
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
		return bits
	} else if encodingType["name"] == "timeBin" {
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
	}
	return bits
}

func (node *Node) setBases(basisList []int, startTime uint64, frequency float64, detectorName string) {
	encodingType := node.components[detectorName].(QSDetector).encodingType
	basisStartTime := startTime - uint64(math.Pow10(12)/(2*frequency))
	if encodingType["name"] == "polarization" {
		splitter := node.components[detectorName].(QSDetector).splitter
		splitter.startTime = basisStartTime
		splitter.frequency = frequency
		var splitterBasisList []*Basis
		for _, d := range basisList {
			splitterBasisList = append(splitterBasisList, encodingType["bases"].([]*Basis)[d])
			splitter.basisList = splitterBasisList
		}
	} else if encodingType["name"] == "timeBin" {
		_switch := node.components[detectorName].(QSDetector)._switch
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
	node.components[channel].(ClassicalChannel).transmit(msg, node)
}

func (node *Node) receiveMessage(message kernel.Message) {
	node.message = message
	node.protocol.(BB84).receivedMessage()
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
		results[0] = value
	}
	return results
}

func outer(a, b []complex128, conj bool) *Basis { // assume a and b are n*1 and 1*m matrix
	result := make(Basis, len(a))
	if conj {
		for i, c := range a {
			for _, d := range b {
				result[i] = append(result[i], cmplx.Conj(c)*(d))
			}
		}
	} else {
		for i, c := range a {
			for _, d := range b {
				result[i] = append(result[i], (c)*(d))
			}
		}
	}
	return &result
}

//func kron(a,b *Basis) *Basis{
//	result := make(Basis,len(*a)*len(*b))
//	for i:= 0; i<len(*a); i++ {
//		for i1:=0; i1<len((*a)[0]);i1++{
//			for j:= 0; j<len(*b); j++  {
//				for j1:= 0; j1 < len((*b)[0]); j++{
//					result[i+j] = append(result[i+j],(*a)[i][i1]*(*b)[j][j1])
//				}
//			}
//		}
//	}
//	return &result
//}

func transpose(a *Basis, conj bool) *Basis {
	result := make(Basis, len((*a)[0]))
	if conj {
		for i := 0; i < len(*a); i++ {
			for j := 0; j < len((*a)[i]); j++ {
				result[j] = append(result[j], cmplx.Conj((*a)[i][j]))
			}
		}
	} else {
		for i := 0; i < len(*a); i++ {
			for j := 0; j < len((*a)[i]); j++ {
				result[j] = append(result[j], (*a)[i][j])
			}
		}
	}
	return &result
}

func matMul(a, b *Basis) *Basis { // Matrix multiplication
	if len((*a)[0]) != len(*b) {
		panic("the columns of first matrix must equal to the rows of the second matrix")
	}
	result := make(Basis, len(*a))
	for i := 0; i < len(result); i++ {
		for j := 0; j < len(*b); j++ {
			val := helpMatMul(a, b, i, j) //a[i][]*b[][j]
			result[i] = append(result[i], val)
		}
	}
	return &result
}

func helpMatMul(a, b *Basis, aIndex int, bIndex int) complex128 { // a[i][] * b[][j]
	var result complex128
	for i := 0; i < len(*a); i++ {
		result += (*a)[aIndex][i] * (*b)[i][bIndex]
	}
	return result
}

func oneToTwo(a []complex128) *Basis { //one dimension to two dimension array
	result := make(Basis, 1)
	for i := 0; i < len(a); i++ {
		result[0] = append(result[0], a[i])
	}
	return &result
}

func divide(a *Basis, b float64) *Basis {
	result := make(Basis, len(*a))
	for i := 0; i < len(*a); i++ {
		for j := 0; j < len((*a)[0]); j++ {
			result[i] = append(result[i], (*a)[i][j]/complex(b, 0))
		}
	}
	return &result
}
