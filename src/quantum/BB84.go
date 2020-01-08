package quantum

import (
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

type BB84 struct {
	name           string
	detectorName   string
	sourceName     string
	working        bool
	ready          bool
	lightTime      float64
	qubitFrequency float64
	startTime      uint64
	classicalDelay float64
	quantumDelay   int
	photonDelay    int
	basicLists     [][]int
	bitLists       [][]int
	key            int64
	keyBits        []int
	node           *Node
	parent         *Parent
	another        *BB84
	keyLength      []int
	keysLeftList   []int
	endRunTimes    []uint64
	latency        float64
	lastKeyTime    uint64
	throughPuts    []float64
	errorRates     []float64
	timeline       *kernel.Timeline
	role           int
}

func (bb84 *BB84) assignNode(node *Node) {
	bb84.node = node
	cchannel := node.components["cchannel"].(ClassicalChannel)
	qchannel := node.components["qchannel"].(QuantumChannel)
	bb84.classicalDelay = cchannel.delay
	bb84.quantumDelay = int(math.Round(qchannel.distance / qchannel.lightSpeed))
}

func (bb84 *BB84) addParent(parent *Parent) {
	bb84.parent = parent
}

func (bb84 *BB84) delParent() {
	bb84.parent = nil
}

func (bb84 *BB84) setBases() {
	numPulses := int(math.Round(bb84.lightTime * bb84.qubitFrequency))
	basisList := choice([]int{0, 1}, numPulses)
	bb84.basicLists = append(bb84.basicLists, basisList)
	bb84.node.setBases(basisList, bb84.startTime, bb84.qubitFrequency, bb84.detectorName)
}

func (bb84 *BB84) beginPhotonPulse(message kernel.Message) {
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		// generate basis/bit list
		numPulses := int(math.Round(bb84.lightTime * bb84.qubitFrequency))
		basisList := choice([]int{0, 1}, numPulses)
		bitList := choice([]int{0, 1}, numPulses)
		// emit photons
		bb84.node.sendQubits(basisList, bitList, bb84.sourceName)
		bb84.basicLists = append(bb84.basicLists, basisList)
		bb84.bitLists = append(bb84.bitLists, bitList)
		// schedule another
		bb84.startTime = bb84.timeline.Now()
		message := kernel.Message{}
		process := kernel.Process{Fnptr: bb84.beginPhotonPulse, Message: message, Owner: bb84.timeline}
		time := bb84.startTime + uint64(math.Round(bb84.lightTime*math.Pow10(12)))
		event := kernel.Event{Time: time, Process: &process, Priority: 0}
		bb84.timeline.Schedule(&event)
	} else {
		bb84.working = false
		bb84.another.working = false
		bb84.keyLength = bb84.keyLength[1:]
		bb84.keysLeftList = bb84.keysLeftList[1:]
		bb84.another.endRunTimes = bb84.another.endRunTimes[1:]
		bb84.another.keyLength = bb84.another.keyLength[1:]
		bb84.another.endRunTimes = bb84.another.endRunTimes[1:]
		// wait for quantum channel to clear of photons, then start protocol
		time := bb84.timeline.Now() + uint64(bb84.quantumDelay+1)
		message := kernel.Message{}
		process := kernel.Process{Fnptr: bb84.startProtocol, Message: message, Owner: bb84.timeline}
		event := kernel.Event{Time: time, Process: &process, Priority: 0}
		bb84.timeline.Schedule(&event)
	}
}

func (bb84 *BB84) endPhotonPulse(message kernel.Message) {
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		// get bits
		bb84.bitLists = append(bb84.bitLists, bb84.node.getBits(bb84.lightTime, bb84.startTime, bb84.qubitFrequency, bb84.detectorName)) // upadate later
		// clear detector photon times to restart measurement
		bb84.node.components[bb84.detectorName].(QSDetector).clearDetectors(kernel.Message{})
		// schedule another if necessary
		if bb84.timeline.Now()+uint64(math.Round(bb84.lightTime*math.Pow10(12))) < bb84.endRunTimes[0] {
			bb84.startTime = bb84.timeline.Now()
			// set bases for measurement
			bb84.setBases()
			// schedule another
			time := bb84.timeline.Now() + uint64(bb84.quantumDelay+1)
			message := kernel.Message{}
			process := kernel.Process{Fnptr: bb84.endPhotonPulse, Message: message, Owner: bb84.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			bb84.timeline.Schedule(&event)
		}
		bb84.node.sendMessage("receivedQubits", "cchannel")
	}
}

func (bb84 *BB84) receivedMessage() {
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		message0 := strings.Split(bb84.node.message["message"].(string), " ")
		if message0[0] == "beginPhotonPulse" {
			bb84.qubitFrequency, _ = strconv.ParseFloat(message0[1], 64)
			bb84.lightTime, _ = strconv.ParseFloat(message0[2], 64)
			bb84.startTime, _ = strconv.ParseUint(message0[3], 10, 64)
			bb84.startTime += uint64(bb84.quantumDelay)
			//generate basis list and set bases for measurement
			bb84.setBases()
			//schedule end_photon_pulse()
			message := kernel.Message{}
			process := kernel.Process{Fnptr: bb84.endPhotonPulse, Message: message, Owner: bb84.timeline}
			time := bb84.startTime + uint64(math.Round(bb84.lightTime*math.Pow10(12)))
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			bb84.timeline.Schedule(&event)
			//clear detector photon times to restart measurement
			process = kernel.Process{Fnptr: bb84.node.components[bb84.detectorName].(QSDetector).clearDetectors, Message: message, Owner: bb84.timeline}
			event = kernel.Event{Time: bb84.startTime, Process: &process, Priority: 0}
			bb84.timeline.Schedule(&event)
		} else if message0[0] == "receivedQubits" {
			bases := bb84.basicLists[0]
			bb84.basicLists = bb84.basicLists[1:]
			bb84.node.sendMessage("basisList "+toString(bases), "cchannel") // need to do
		} else if message0[0] == "basisList" {
			var basisListAlice []int
			for _, basis := range message0[1:] {
				value, _ := strconv.Atoi(basis)
				basisListAlice = append(basisListAlice, value)
			}
			var indices []int
			basisList := bb84.basicLists[0]
			bits := bb84.bitLists[0]
			bb84.basicLists = bb84.basicLists[1:]
			bb84.bitLists = bb84.bitLists[1:]
			for i, b := range basisListAlice {
				if bits[i] != -1 && basisList[i] == b {
					indices = append(indices, i)
					bb84.keyBits = append(bb84.keyBits, bits[i])
				}
			}
			bb84.node.sendMessage("matchingIndices "+toString(indices), "cchannel")
		} else if message0[0] == "matchingIndices" {
			// need to do
			var indices []int
			if len(message0) != 1 {
				for _, val := range message0[1:] {
					v, _ := strconv.Atoi(val)
					indices = append(indices, v)
				}
			}
			bits := bb84.bitLists[0]
			bb84.bitLists = bb84.bitLists[1:]
			for i := range indices {
				bb84.keyBits = append(bb84.keyBits, bits[i])
			}
			if len(bb84.keyBits) >= bb84.keyLength[0] {
				throughput := float64(bb84.keyLength[0]) * math.Pow10(12) / float64(bb84.timeline.Now()-bb84.lastKeyTime)
				for len(bb84.keyBits) >= bb84.keyLength[0] && bb84.keysLeftList[0] > 0 {
					fmt.Println("got key")
					bb84.setKey()
					if bb84.parent != nil {
						bb84.parent.getKeyFromBB84(bb84.key)
					}
					bb84.another.setKey()
					if bb84.another.parent != nil {
						bb84.another.parent.getKeyFromBB84(bb84.key)
					}
					// for metrics
					if bb84.latency == 0 {
						bb84.latency = float64(bb84.timeline.Now()-bb84.lastKeyTime) * math.Pow10(-12)
					}
					bb84.throughPuts = append(bb84.throughPuts, throughput)
					keyDiff := bb84.key ^ bb84.another.key
					numErrors := 0
					for keyDiff != 0 {
						keyDiff &= keyDiff - 1
						numErrors += 1
					}
					bb84.errorRates = append(bb84.errorRates, float64(numErrors)/float64(bb84.keyLength[0]))
					bb84.keysLeftList[0] -= 1
				}
				bb84.lastKeyTime = bb84.timeline.Now()
			}
			if bb84.keysLeftList[0] < 1 {
				bb84.working = false
				bb84.another.working = false
			}
		}
	}
}

func (bb84 *BB84) generateKey(length, keyNum int, runTime uint64) {
	if bb84.role != 0 { // 0: Alice 1:Bob
		panic("generate key must be called from Alice")
	}
	bb84.keyLength = append(bb84.keyLength, length)
	bb84.another.keyLength = append(bb84.another.keyLength, length)
	bb84.keysLeftList = append(bb84.keysLeftList, keyNum)
	endRunTime := runTime + bb84.timeline.Now()
	bb84.endRunTimes = append(bb84.endRunTimes, endRunTime)
	bb84.another.endRunTimes = append(bb84.another.endRunTimes, endRunTime)
	if bb84.ready {
		bb84.ready = false
		bb84.working = true
		bb84.another.working = true
		bb84.startProtocol(kernel.Message{})
	}
}

func (bb84 *BB84) startProtocol(message kernel.Message) {
	if len(bb84.keyLength) > 0 {
		bb84.basicLists = [][]int{}
		bb84.another.basicLists = [][]int{}
		bb84.bitLists = [][]int{}
		bb84.another.bitLists = [][]int{}
		bb84.keyBits = []int{}
		bb84.another.keyBits = []int{}
		bb84.latency = 0
		bb84.another.latency = 0
		bb84.working = true
		bb84.another.working = true
		// turn on bob's detectors
		bb84.another.node.components[bb84.another.detectorName].(QSDetector).turnOnDetectors()

		lightSource := bb84.node.components[bb84.sourceName].(LightSource)
		bb84.qubitFrequency = lightSource.frequency
		// calculate light time based on key length
		bb84.lightTime = float64(bb84.keyLength[0]) / (bb84.qubitFrequency * lightSource.meanPhotonNum)
		// send message that photon pulse is beginning, then send bits
		bb84.startTime = (bb84.timeline.Now()) + uint64(math.Round(bb84.classicalDelay))
		bb84.node.sendMessage("beginPhotonPulse "+fmt.Sprint(bb84.qubitFrequency)+
			" "+fmt.Sprint(bb84.lightTime)+" "+fmt.Sprint(bb84.startTime)+" "+
			fmt.Sprint(lightSource.wavelength), "cchannel")

		message := kernel.Message{}
		process := kernel.Process{Fnptr: bb84.beginPhotonPulse, Message: message, Owner: bb84.timeline}
		event := kernel.Event{Time: bb84.startTime, Process: &process, Priority: 0}
		bb84.timeline.Schedule(&event)
		bb84.lastKeyTime = bb84.timeline.Now()
	} else {
		bb84.another.node.components[bb84.another.detectorName].(QSDetector).turnOffDetectors()
		bb84.ready = true
	}
}

func (bb84 *BB84) setKey() {
	keyBits := bb84.keyBits[0:bb84.keyLength[0]]
	bb84.keyBits = bb84.keyBits[bb84.keyLength[0]:]
	bb84.key = sliceToString(keyBits)
}

func choice(array []int, n int) []int {
	basisList := make([]int, n)
	for i := 0; i < n; i++ {
		basisList[i] = array[rand.Intn(len(array))]
	}
	return basisList
}

func sliceToString(slice []int) int64 { // convert from binary list to int
	var results string
	for i := range slice {
		results += strconv.Itoa(slice[i])
	}
	val, _ := strconv.ParseInt(results, 2, 64)
	return val
}

func toString(a []int) string {
	valuesText := make([]string, len(a))
	for i := 0; i < len(a); i++ {
		valuesText[i] = strconv.Itoa(a[i])
	}
	result := strings.Join(valuesText, " ")
	return result
}

// for testing BB84 Protocol
type Parent struct {
	keySize int
	role    int
	child   *BB84
	key     int64
}

func (parent *Parent) run(message kernel.Message) {
	parent.child.generateKey(parent.keySize, 10, math.MaxInt64)
}

func (parent *Parent) getKeyFromBB84(key int64) {
	fmt.Println("need to do")
	parent.key = key
}

func test() {
	fmt.Println("Polarization:")
	tl := kernel.Timeline{Name: "alice", LookAhead: math.MaxInt64}
	tl.SetEndTime(uint64(math.Pow10(11))) //stop time is 100 ms
	op := OpticalChannel{lightSpeed: 3 * math.Pow10(-4), polarizationFidelity: 0.99, attenuation: 0.0002, distance: math.Pow10(3)}
	qc := QuantumChannel{name: "qc", timeline: &tl, OpticalChannel: op}
	cc := ClassicalChannel{name: "cc", timeline: &tl, OpticalChannel: op}
	// Alice
	ls := LightSource{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc}
	components := map[string]interface{}{"lightsource": ls, "cchannel": cc, "qchannel": qc}
	alice := Node{name: "alice", timeline: &tl, components: components}
	qc.setSender(&ls)
	cc.addEnd(&alice)

	//Bob
	detectors := []Detector{{efficiency: 0.8, darkCount: 1, timeResolution: 10}, {efficiency: 0.8, darkCount: 1, timeResolution: 10}}
	qsd := QSDetector{name: "bob.qsdetector", timeline: &tl, detectors: detectors}
	qsd._init()
	components = map[string]interface{}{"detector": qsd, "cchannel": cc, "qchannel": qc}
	bob := Node{name: "bob", timeline: &tl, components: components}
	qc.setReceiver(&qsd)
	cc.addEnd(&bob)

	// init() components elements
	// need to do

	//BB84
	bba := BB84{name: "bba", timeline: &tl, role: 0} //alice.role = 0
	bbb := BB84{name: "bbb", timeline: &tl, role: 1} //bob.role = 1
	bba.assignNode(&alice)
	bbb.assignNode(&bob)
	bba.another = &bbb
	bbb.another = &bba
	alice.protocol = bba
	bob.protocol = bbb

	//Parent
	pa := Parent{keySize: 512}
	pb := Parent{keySize: 512}
	pa.child = &bba
	pb.child = &bbb
	bba.addParent(&pa)
	bbb.addParent(&pb)

	message := kernel.Message{}
	process := kernel.Process{Fnptr: pa.run, Message: message, Owner: &tl}
	event := kernel.Event{Time: 0, Priority: 0, Process: &process}
	tl.Schedule(&event)
	kernel.Run([]*kernel.Timeline{&tl})

	fmt.Println("latency (s): " + fmt.Sprintf("%f", bba.latency))
	//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
	fmt.Println("bit error rates:")

	// TIME BIN TESTING
	fmt.Println("Time Bin:")
	tl = kernel.Timeline{Name: "alice", LookAhead: math.MaxInt64}
	tl.SetEndTime(uint64(math.Pow10(11))) //stop time is 100 ms
	op = OpticalChannel{lightSpeed: 3 * math.Pow10(-4), polarizationFidelity: 0.99, distance: math.Pow10(3)}
	qc = QuantumChannel{name: "qc", timeline: &tl, OpticalChannel: op}
	cc = ClassicalChannel{name: "cc", timeline: &tl, OpticalChannel: op}
	// Alice
	ls = LightSource{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, encodingType: timeBin()}
	components = map[string]interface{}{"asource": ls, "cchannel": cc, "qchannel": qc}
	alice = Node{name: "alice", timeline: &tl, components: components}
	qc.setSender(&ls)
	cc.addEnd(&alice)

	//Bob
	detectors = []Detector{{efficiency: 0.8, darkCount: 1, timeResolution: 10}, {efficiency: 0.8, darkCount: 1, timeResolution: 10}, {efficiency: 0.8, darkCount: 1, timeResolution: 10}}
	interferometer := Interferometer{pathDifference: timeBin()["binSeparation"].(int)}
	qsd = QSDetector{name: "bob.qsdetector", timeline: &tl, detectors: detectors, encodingType: timeBin(), interferometer: &interferometer}
	qsd._init()
	components = map[string]interface{}{"bdetector": qsd, "cchannel": cc, "qchannel": qc}
	bob = Node{name: "bob", timeline: &tl, components: components}
	qc.setReceiver(&qsd)
	cc.addEnd(&bob)

	// init() components elements
	// need to do

	//BB84
	bba = BB84{name: "bba", timeline: &tl, role: 0, sourceName: "asource"}   //alice.role = 0
	bbb = BB84{name: "bbb", timeline: &tl, role: 1, sourceName: "bdetector"} //bob.role = 1
	bba.assignNode(&alice)
	bbb.assignNode(&bob)
	bba.another = &bbb
	bbb.another = &bba
	alice.protocol = bba
	bob.protocol = bbb

	//Parent
	pa = Parent{keySize: 512}
	pb = Parent{keySize: 512}
	pa.child = &bba
	pb.child = &bbb
	bba.addParent(&pa)
	bbb.addParent(&pb)

	message = kernel.Message{}
	process = kernel.Process{Fnptr: pa.run, Message: message, Owner: &tl}
	event = kernel.Event{Time: 0, Priority: 0, Process: &process}
	tl.Schedule(&event)
	kernel.Run([]*kernel.Timeline{&tl})

	fmt.Println("latency (s): " + fmt.Sprintf("%f", bba.latency))
	//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
	fmt.Println("bit error rates:")
}
