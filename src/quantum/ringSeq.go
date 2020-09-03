package quantum

import (
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
	"sync"
)

func Main(n int, threadNum int, lookAhead uint64) {
	fmt.Println("Ring QKD Network (sequential)")

	tls := make([]*kernel.Timeline, threadNum)
	nodeOnThread := n / threadNum
	count := 0
	photonNumber := 0
	for i := 0; i < threadNum; i++ {
		tlName := fmt.Sprint("timeline", i)
		tl := kernel.Timeline{Name: tlName}
		tl.Init(lookAhead, uint64(math.Pow10(10))) // 10 ms
		var eventPool = sync.Pool{
			New: func() interface{} {
				message := &kernel.Message{}
				process := &kernel.Process{}
				process.Message = *message
				event := &kernel.Event{}
				event.Process = process
				event.Priority = 0
				count++
				return event
			},
		}

		var photonPool = sync.Pool{
			New: func() interface{} {
				photon := &Photon{}
				photonNumber++
				return photon
			},
		}
		tl.EventPool = &eventPool
		tl.PhotonPool = &photonPool

		tls[i] = &tl
	}

	// create nodes
	totalNodes := n
	nodes := make([]*Node, totalNodes)

	for i := 0; i < totalNodes; i++ {
		nodeName := fmt.Sprint("node", i)
		node := Node{name: nodeName, timeline: tls[i/nodeOnThread]}
		node.cchannels = make(map[string]*ClassicalChannel)
		node.components = make(map[string]interface{})
		nodes[i] = &node
	}

	// create classical channels

	for i := 0; i < totalNodes; i++ {
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		ccName := fmt.Sprint("cc_", nodes[i].name, "_", nodes[(i+1)%totalNodes].name)
		cc := &ClassicalChannel{name: ccName, OpticalChannel: op, delay: 1 * math.Pow10(9)}
		cc.SetSender(nodes[i])
		cc.SetReceiver(nodes[(i+1)%totalNodes])
		nodes[i].assignCChannel(cc)

		ccName = fmt.Sprint("cc_", nodes[(i+1)%totalNodes].name, "_", nodes[i].name)
		cc = &ClassicalChannel{name: ccName, OpticalChannel: op, delay: 1 * math.Pow10(9)}
		cc.SetSender(nodes[(i+1)%totalNodes])
		cc.SetReceiver(nodes[i])
		nodes[(i+1)%totalNodes].assignCChannel(cc)
	}

	// create light source, detector and quantum channels
	for i := 0; i < totalNodes; i++ {
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		qcName := fmt.Sprint("qc_", nodes[i].name, "_", nodes[(i+1)%totalNodes].name)
		qc := QuantumChannel{name: qcName, timeline: tls[i/nodeOnThread], OpticalChannel: op}
		qc.init()
		lsName := fmt.Sprint(nodes[i].name, ".lightsource")
		ls := LightSource{name: lsName, timeline: tls[i/nodeOnThread], frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, wavelength: 1550, encodingType: polarization()}
		ls.init()
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}}
		qsdName := fmt.Sprint(nodes[(i+1)%totalNodes].name, ".qsdetector")
		qsd := QSDetector{name: qsdName, timeline: tls[((i+1)%totalNodes)/nodeOnThread], detectors: detectors}
		qc.setReceiver(&qsd)
		nodes[i].components["lightSource"] = &ls
		qsd._init()
		qsd.init()
		nodes[(i+1)%totalNodes].components["detector"] = &qsd
	}

	// create BB84
	parent_protocols := make([]*Parent, 0)
	for i := 0; i < totalNodes; i++ {
		bbName := fmt.Sprint(nodes[i].name, ".bba")
		bba := BB84{name: bbName, timeline: tls[i/nodeOnThread], role: 0} //alice.role = 0
		bbName = fmt.Sprint(nodes[(i+1)%totalNodes].name, ".bbb")
		bbb := BB84{name: bbName, timeline: tls[((i+1)%totalNodes)/nodeOnThread], role: 1} //bob.role = 1
		bba._init()
		bbb._init()
		bba.assignNode(nodes[i], 1*math.Pow10(9), 50000000)
		bbb.assignNode(nodes[(i+1)%totalNodes], 1*math.Pow10(9), 50000000)
		bba.another = &bbb
		bbb.another = &bba
		// TODO: assign protocols to nodes
		pa := Parent{keySize: 512, role: "alice"}
		parent_protocols = append(parent_protocols, &pa)
		pb := Parent{keySize: 512, role: "bob"}
		pa.child = &bba
		pb.child = &bbb
		bba.addParent(&pa)
		bbb.addParent(&pb)
	}

	// schedule initial events
	for i := 0; i < len(parent_protocols); i++ {

		event := tls[i/nodeOnThread].EventPool.Get().(*kernel.Event)
		event.Time = 0
		event.Priority = 0
		event.Process.Fnptr = parent_protocols[i].run
		event.Process.Owner = tls[i/nodeOnThread]
		tls[i/nodeOnThread].Schedule(event)

		//message := kernel.Message{}
		//process := kernel.Process{Fnptr: parent_protocols[i].run, Message: message, Owner: tls[i/nodeOnThread]}
		//event := kernel.Event{Time: 0, Priority: 0, Process: &process}
		//tls[i/nodeOnThread].Schedule(&event)
	}

	kernel.Run(tls)

	//for i := 0; i < totalNodes; i++ {
	//	fmt.Println(nodes[i].name)
	//	fmt.Println("   latency (s): " + fmt.Sprintf("%f", parent_protocols[i].child.latency))
	//	//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
	//	fmt.Print("   average throughput (Mb/s): ")
	//	fmt.Println("   ", 1e-6*floats.Sum(parent_protocols[i].child.throughPuts)/float64(len(parent_protocols[i].child.errorRates)))
	//	//fmt.Println("   bit error rates:")
	//	//for i, e := range parent_protocols[i].child.errorRates {
	//	//	fmt.Println("\tkey " + strconv.Itoa(i+1) + ":\t" + fmt.Sprintf("%f", e*100) + "%")
	//	//}
	//	fmt.Print("   sum error rates: ")
	//	fmt.Println("   ", floats.Sum(parent_protocols[i].child.errorRates)/float64(len(parent_protocols[i].child.errorRates)))
	//}

	fmt.Println("sync counter:", tls[0].SyncCounter)
	fmt.Println("end time", tls[0].Now())
	fmt.Println("Created: ", count)
}
