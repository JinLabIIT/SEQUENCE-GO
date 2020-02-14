package quantum

import (
	"github.com/gonum/floats"
	"github.com/leesper/go_rng"
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
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
	basisLists     [][]int
	bitLists       [][]int
	key            []uint64
	combine        int //we assume key size = n * 64, combine = keySize / 64
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
	rng            *rng.UniformGenerator
}

func (bb84 *BB84) _init() {
	bb84.rng = rng.NewUniformGenerator(123)
	bb84.ready = true
	if bb84.sourceName == "" {
		bb84.sourceName = "lightSource"
	}
	if bb84.detectorName == "" {
		bb84.detectorName = "detector"
	}
}

func (bb84 *BB84) assignNode(node *Node, cdelay float64, qdelay int) {
	bb84.node = node
	//cchannel := node.components["cchannel"].(*ClassicalChannel)
	//qchannel := node.components["qchannel"].(*QuantumChannel)
	//bb84.classicalDelay = cchannel.delay
	//bb84.quantumDelay = int(math.Round(qchannel.distance / qchannel.lightSpeed))
	bb84.quantumDelay = qdelay
	bb84.classicalDelay = cdelay
	node.protocols = append(node.protocols, bb84)
}

func (bb84 *BB84) addParent(parent *Parent) {
	bb84.parent = parent
}

func (bb84 *BB84) delParent() {
	bb84.parent = nil
}

func (bb84 *BB84) setBases() {
	numPulses := int(math.Round(bb84.lightTime * bb84.qubitFrequency))
	basisList := choice([]int{0, 1}, numPulses, bb84.rng) //create an numPulses length array and the element is randomly chosen between 1 and 0
	bb84.basisLists = append(bb84.basisLists, basisList)
	bb84.node.setBases(basisList, bb84.startTime, bb84.qubitFrequency, bb84.detectorName)
}

func (bb84 *BB84) beginPhotonPulse(message kernel.Message) {
	// fmt.Println("beginPhotonPulse " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10))
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		// generate basis/bit list
		numPulses := int(math.Round(bb84.lightTime * bb84.qubitFrequency))
		basisList := choice([]int{0, 1}, numPulses, bb84.rng)
		bitList := choice([]int{0, 1}, numPulses, bb84.rng)
		// emit photons
		bb84.node.sendQubits(basisList, bitList, bb84.sourceName)
		bb84.basisLists = append(bb84.basisLists, basisList)
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
		bb84.endRunTimes = bb84.endRunTimes[1:]
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
	// fmt.Println("endPhotonPulse " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10))
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		// get bits
		bits := bb84.node.getBits(bb84.lightTime, bb84.startTime, bb84.qubitFrequency, bb84.detectorName)
		bb84.bitLists = append(bb84.bitLists, bits) // upadate later
		// clear detector photon times to restart measurement
		bb84.node.components[bb84.detectorName].(*QSDetector).clearDetectors(kernel.Message{})
		// schedule another if necessary
		if bb84.timeline.Now()+uint64(math.Round(bb84.lightTime*math.Pow10(12))) < bb84.endRunTimes[0] {
			bb84.startTime = bb84.timeline.Now()
			// set bases for measurement
			bb84.setBases()
			// schedule another
			time := bb84.startTime + uint64(math.Round(bb84.lightTime*math.Pow10(12)))
			message := kernel.Message{}
			process := kernel.Process{Fnptr: bb84.endPhotonPulse, Message: message, Owner: bb84.timeline}
			event := kernel.Event{Time: time, Process: &process, Priority: 0}
			bb84.timeline.Schedule(&event)
		}
		bb84.node.sendMessage("receivedQubits", bb84.another.node.name)
	}
}

func (bb84 *BB84) receivedMessage(message kernel.Message) {
	if bb84.another.node.name != message["src"] {
		return
	}
	if bb84.working && bb84.timeline.Now() < bb84.endRunTimes[0] {
		message0 := strings.Split(message["message"].(string), " ")
		if message0[0] == "beginPhotonPulse" {
			if bb84.role != 1 {
				return
			}
			//fmt.Println("beginPhotonPulse in received message " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10))
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
			process1 := kernel.Process{Fnptr: bb84.node.components[bb84.detectorName].(*QSDetector).clearDetectors, Message: message, Owner: bb84.timeline}
			event1 := kernel.Event{Time: bb84.startTime, Process: &process1, Priority: 0}
			bb84.timeline.Schedule(&event1)
		} else if message0[0] == "receivedQubits" {
			if bb84.role != 0 {
				return
			}
			//fmt.Println("receivedQubits " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10))
			bases := bb84.basisLists[0]
			bb84.basisLists = bb84.basisLists[1:]
			bb84.node.sendMessage("basisList "+toString(bases), bb84.another.node.name) // need to do
		} else if message0[0] == "basisList" {
			if bb84.role != 1 {
				return
			}
			//fmt.Println("basislist " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10), "from", message["src"])
			basisListAlice := make([]int, 0, len(message0[1:]))
			for _, basis := range message0[1:] {
				value, _ := strconv.Atoi(basis)
				basisListAlice = append(basisListAlice, value)
			}
			indices := make([]int, 0, 200)
			basisList := bb84.basisLists[0]
			bits := bb84.bitLists[0]
			bb84.basisLists = bb84.basisLists[1:]
			bb84.bitLists = bb84.bitLists[1:]
			a := 0 // need to be deleted
			c := 0 // need to be deleted
			for i, b := range basisListAlice {
				if bits[i] != -1 && basisList[i] == b {
					indices = append(indices, i)
					bb84.keyBits = append(bb84.keyBits, bits[i])
				}
				if bits[i] == -1 {
					a++
				}
				if basisList[i] == b {
					c++
				}
			}
			bb84.node.sendMessage("matchingIndices "+toString(indices), bb84.another.node.name)
		} else if message0[0] == "matchingIndices" {
			// need to do
			if bb84.role != 0 {
				return
			}
			// fmt.Println("matchingIndices " + bb84.name + " " + strconv.FormatUint(bb84.timeline.Now(), 10))
			indices := make([]int, 0, len(message0[1:]))
			if len(message0) != 1 { // no matching indices
				for _, val := range message0[1:] {
					v, _ := strconv.Atoi(val)
					indices = append(indices, v)
				}
			}
			bits := bb84.bitLists[0]
			bb84.bitLists = bb84.bitLists[1:]
			for _, i := range indices {
				bb84.keyBits = append(bb84.keyBits, bits[i])
			}
			if len(bb84.keyBits) >= bb84.keyLength[0] {
				throughput := float64(bb84.keyLength[0]) * math.Pow10(12) / float64(bb84.timeline.Now()-bb84.lastKeyTime)
				for len(bb84.keyBits) >= bb84.keyLength[0] && bb84.keysLeftList[0] > 0 {
					// fmt.Println(bb84.node.name, "got key")
					bb84.setKey()
					if bb84.parent != nil {
						bb84.parent.getKeyFromBB84(bb84.key)
					}
					bb84.another.setKey()
					if bb84.another.parent != nil {
						bb84.another.parent.getKeyFromBB84(bb84.another.key)
					}
					// for metrics
					if bb84.latency == 0 {
						bb84.latency = float64(bb84.timeline.Now()-bb84.lastKeyTime) * math.Pow10(-12)
					}
					bb84.throughPuts = append(bb84.throughPuts, throughput)
					keyDiff := make([]uint64, bb84.combine)
					for i, val := range bb84.key {
						keyDiff[i] = val ^ bb84.another.key[i]
					}
					//keyDiff := bb84.key ^ bb84.another.key
					numErrors := 0
					for i := range keyDiff {
						val := keyDiff[i]
						for val != 0 {
							val &= val - 1
							numErrors += 1
						}
					}
					/*for keyDiff != 0 {
						keyDiff &= keyDiff - 1
						numErrors += 1
					}*/
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
	//fmt.Println("generateKey " + bb84.name)
	if bb84.role != 0 { // 0: Alice 1:Bob
		panic("generate key must be called from Alice")
	}
	bb84.keyLength = append(bb84.keyLength, length)
	bb84.another.keyLength = append(bb84.another.keyLength, length)
	bb84.keysLeftList = append(bb84.keysLeftList, keyNum)
	endRunTime := runTime + bb84.timeline.Now()
	bb84.endRunTimes = append(bb84.endRunTimes, endRunTime)
	bb84.another.endRunTimes = append(bb84.another.endRunTimes, endRunTime)

	bb84.combine = bb84.keyLength[0] / 64
	bb84.another.combine = bb84.combine
	bb84.key = make([]uint64, bb84.combine)
	bb84.another.key = make([]uint64, bb84.combine)
	if bb84.ready {
		bb84.ready = false
		bb84.working = true
		bb84.another.working = true
		bb84.startProtocol(kernel.Message{})
	}
}

func (bb84 *BB84) startProtocol(message kernel.Message) {
	//fmt.Println("startProtocol " + bb84.name)
	if len(bb84.keyLength) > 0 {
		bb84.basisLists = [][]int{}
		bb84.another.basisLists = [][]int{}
		bb84.bitLists = [][]int{}
		bb84.another.bitLists = [][]int{}
		bb84.keyBits = []int{}
		bb84.another.keyBits = []int{}
		bb84.latency = 0
		bb84.another.latency = 0

		bb84.working = true
		bb84.another.working = true
		// turn on bob's detectors
		bb84.another.node.components[bb84.another.detectorName].(*QSDetector).turnOnDetectors()

		lightSource := bb84.node.components[bb84.sourceName].(*LightSource)
		bb84.qubitFrequency = lightSource.frequency
		// calculate light time based on key length
		bb84.lightTime = float64(bb84.keyLength[0]) / (bb84.qubitFrequency * lightSource.meanPhotonNum)
		// send message that photon pulse is beginning, then send bits
		bb84.startTime = (bb84.timeline.Now()) + uint64(math.Round(bb84.classicalDelay))
		bb84.node.sendMessage("beginPhotonPulse "+fmt.Sprint(bb84.qubitFrequency)+
			" "+fmt.Sprint(bb84.lightTime)+" "+fmt.Sprint(bb84.startTime)+" "+
			fmt.Sprint(lightSource.wavelength), bb84.another.node.name)

		message := kernel.Message{}
		process := kernel.Process{Fnptr: bb84.beginPhotonPulse, Message: message, Owner: bb84.timeline}
		event := kernel.Event{Time: bb84.startTime, Process: &process, Priority: 0}
		bb84.timeline.Schedule(&event)

		bb84.lastKeyTime = bb84.timeline.Now()
	} else {
		bb84.another.node.components[bb84.another.detectorName].(*QSDetector).turnOffDetectors()
		bb84.ready = true
	}
}

func (bb84 *BB84) setKey() {
	//keyBits := bb84.keyBits[0:bb84.keyLength[0]]
	keyBits := make([]int, 0, 64)
	for i := 0; i < bb84.combine; i++ {
		keyBits = bb84.keyBits[i*64 : (i+1)*64]
		tmp := sliceToInt(keyBits, 2)
		bb84.key[i] = tmp
	}
	if len(bb84.keyBits) >= bb84.combine*64 {
		bb84.keyBits = bb84.keyBits[bb84.combine*64:]
	}
	//bb84.key = sliceToInt(keyBits,2)//convert from binary list to int
}

// for testing BB84 Protocol
type Parent struct {
	keySize int
	role    string
	child   *BB84
	key     []uint64
}

func (parent *Parent) run(message kernel.Message) {
	//parent.combine = keySize / 64
	parent.child.generateKey(parent.keySize, 1000000, math.MaxInt64)
}

func (parent *Parent) getKeyFromBB84(key []uint64) {
	// fmt.Print("key for " + parent.role + ": ")
	var str string
	for _, val := range key {
		str += strconv.FormatUint(val, 2)
	}
	// fmt.Println(str)
	parent.key = key
}

func test() {
	seed := int64(156)
	fmt.Println("Polarization:")
	poisson := rng.NewPoissonGenerator(seed)
	tl := kernel.Timeline{Name: "timeline", LookAhead: math.MaxInt64}
	tl.SetEndTime(uint64(math.Pow10(11))) //stop time is 100 ms
	op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
	qc := QuantumChannel{name: "qc", timeline: &tl, OpticalChannel: op}
	// Alice
	ls := LightSource{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, poisson: poisson, wavelength: 1550, encodingType: polarization()}
	components := map[string]interface{}{"lightSource": &ls}
	alice := Node{name: "alice", timeline: &tl, components: components}
	qc.setSender(&ls)

	//Bob
	detectors := []*Detector{{efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}}

	qsd := QSDetector{name: "bob.qsdetector", timeline: &tl, detectors: detectors}
	qsd._init()
	qsd.init()

	components = map[string]interface{}{"detector": &qsd}
	bob := Node{name: "bob", timeline: &tl, components: components}
	alice.cchannels = make(map[string]*ClassicalChannel)
	bob.cchannels = make(map[string]*ClassicalChannel)
	qc.setReceiver(&qsd)

	cca := ClassicalChannel{name: "alice_bob", timeline: &tl, OpticalChannel: op, delay: float64(1 * math.Pow10(9))}
	cca.SetSender(&alice)
	cca.SetReceiver(&bob)
	alice.assignCChannel(&cca)

	ccb := ClassicalChannel{name: "bob_alice", timeline: &tl, OpticalChannel: op, delay: float64(1 * math.Pow10(9))}
	ccb.SetSender(&bob)
	ccb.SetReceiver(&alice)
	bob.assignCChannel(&ccb)

	//BB84
	bba := BB84{name: "bba", timeline: &tl, role: 0} //alice.role = 0
	bbb := BB84{name: "bbb", timeline: &tl, role: 1} //bob.role = 1
	bba._init()
	bbb._init()
	bba.assignNode(&alice, cca.delay, int(qc.distance/qc.lightSpeed))
	bbb.assignNode(&bob, cca.delay, int(qc.distance/qc.lightSpeed))
	bba.another = &bbb
	bbb.another = &bba

	//Parent
	pa := Parent{keySize: 512, role: "alice"}
	pb := Parent{keySize: 512, role: "bob"}
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
	fmt.Print("average throughput (Mb/s): ")
	fmt.Println(1e-6 * floats.Sum(bba.throughPuts) / float64(len(bba.errorRates)))
	fmt.Println("bit error rates:")
	for i, e := range bba.errorRates {
		fmt.Println("\tkey " + strconv.Itoa(i+1) + ":\t" + fmt.Sprintf("%f", e*100) + "%")
	}
	fmt.Println("sum error rates: ")
	fmt.Print(floats.Sum(bba.errorRates) / float64(len(bba.errorRates)))
	// TIME BIN TESTING need to modify
}

//func test2() {
//	fmt.Println("Time Bin:")
//	tl := kernel.Timeline{Name: "alice2", LookAhead: math.MaxInt64}
//	tl.SetEndTime(uint64(math.Pow10(13))) //stop time is 100 ms
//	op := OpticalChannel{lightSpeed: 3 * math.Pow10(-4), polarizationFidelity: 0.99, distance: math.Pow10(3)}
//	qc := QuantumChannel{name: "qc", timeline: &tl, OpticalChannel: op}
//
//	// Alice
//	ls := LightSource{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, encodingType: timeBin()}
//	components := map[string]interface{}{"asource": &ls}
//	alice := Node{name: "alice", timeline: &tl, components: components}
//	qc.setSender(&ls)
//
//
//
//	//Bob
//	detectors := []*Detector{{efficiency: 0.8, darkCount: 100, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 100, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 100, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 100, timeResolution: 10, countRate: 50 * math.Pow10(6)}}
//
//	interferometer := Interferometer{pathDifference: timeBin()["binSeparation"].(int)}
//	qsd := QSDetector{name: "bob.qsdetector", timeline: &tl, detectors: detectors, encodingType: timeBin(), interferometer: &interferometer}
//	qsd._init()
//	components = map[string]interface{}{"bdetector": &qsd}
//	bob := Node{name: "bob", timeline: &tl, components: components}
//	qc.setReceiver(&qsd)
//
//	cc := ClassicalChannel{name: "alice_bob", timeline: &tl, OpticalChannel: op, delay: float64(1 * math.Pow10(9))}
//	cc.SetSender(&alice)
//	cc.SetReceiver(&bob)
//	alice.assignCChannel(&cc)
//
//	cc = ClassicalChannel{name: "bob_alice", timeline: &tl, OpticalChannel: op, delay: float64(1 * math.Pow10(9))}
//	cc.SetSender(&bob)
//	cc.SetReceiver(&alice)
//	bob.assignCChannel(&cc)
//
//	// init() components elements
//	qsd.init()
//	// need to do
//
//	//BB84
//	bba := BB84{name: "bba", timeline: &tl, role: 0, sourceName: "asource"}     //alice.role = 0
//	bbb := BB84{name: "bbb", timeline: &tl, role: 1, detectorName: "bdetector"} //bob.role = 1
//	bba.assignNode(&alice, cc.delay, int(qc.lightSpeed/qc.distance))
//	bbb.assignNode(&bob, cc.delay, int(qc.lightSpeed/qc.distance))
//	bba.another = &bbb
//	bbb.another = &bba
//
//	//Parent
//	pa := Parent{keySize: 512}
//	pb := Parent{keySize: 512}
//	pa.child = &bba
//	pb.child = &bbb
//	bba.addParent(&pa)
//	bbb.addParent(&pb)
//
//	message := kernel.Message{}
//	process := kernel.Process{Fnptr: pa.run, Message: message, Owner: &tl}
//	event := kernel.Event{Time: 0, Priority: 0, Process: &process}
//	tl.Schedule(&event)
//	kernel.Run([]*kernel.Timeline{&tl})
//
//	fmt.Println("latency (s): " + fmt.Sprintf("%f", bba.latency))
//	//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
//	fmt.Println("bit error rates:")
//}
