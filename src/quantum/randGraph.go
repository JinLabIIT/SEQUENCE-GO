package quantum

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"kernel"
	"os"
)

func randGraph(threadNum int, lookAhead uint64, path string) {
	SEED := uint64(0)
	SIM_TIME := 1e10

	ATTENUATION := 0.0002
	QCFIDELITY := 0.99
	DISTANCE := 1e4
	LIGHTSPEED := 2e-4
	LIGHTSOURCE_FREQUENCY := 8e7
	LIGHTSOURCE_MEAN := 0.1
	CCDELAY := 1e9

	WAVELEN := 1550.0
	DETECTOR_EFFICIENCY := 0.8
	DARKCOUNT := 0.0
	TIME_RESOLUTION := uint64(10)
	COUNT_RATE := 5e7

	KEYSIZE := 512

	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	var result map[string]interface{}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	//byteValue = byteValue
	err = json.Unmarshal(byteValue, &result)

	nodesGraph := result["nodes"].([]interface{})
	n := len(nodesGraph)
	links := result["links"].([]interface{}) // links: {source: int,target: int}

	fmt.Println("n: ", n, "threadNum: ", threadNum, "lookAhead:", lookAhead)

	rand.Seed(SEED)

	tls := make([]*kernel.Timeline, threadNum)
	for i := 0; i < threadNum; i++ {
		tlName := fmt.Sprint("timeline", i)
		tl := kernel.Timeline{Name: tlName}
		tl.Init(lookAhead, uint64(SIM_TIME)) // 10 ms
		tls[i] = &tl
	}

	// create nodes
	totalNodes := n
	nodes := make([]*Node, totalNodes)
	plan := randomSchedule(threadNum, n)
	for threadId := range plan {
		for _, nodeId := range plan[threadId] {
			nodeName := fmt.Sprint("node", nodeId)
			node := Node{name: nodeName, timeline: tls[threadId]}
			node.cchannels = make(map[string]*ClassicalChannel)
			node.components = make(map[string]interface{})
			nodes[nodeId] = &node
		}
	}

	// create classical channels

	for _, v := range links {
		source := int(v.(map[string]interface{})["source"].(float64))
		target := int(v.(map[string]interface{})["target"].(float64))
		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: DISTANCE, lightSpeed: LIGHTSPEED}
		ccName := fmt.Sprint("cc_", nodes[source].name, "_", nodes[target].name)
		cc := &ClassicalChannel{name: ccName, OpticalChannel: op, delay: CCDELAY}
		cc.SetSender(nodes[source])
		cc.SetReceiver(nodes[target])
		nodes[source].assignCChannel(cc)

		ccName = fmt.Sprint("cc_", nodes[target].name, "_", nodes[source].name)
		cc = &ClassicalChannel{name: ccName, OpticalChannel: op, delay: CCDELAY}
		cc.SetSender(nodes[target])
		cc.SetReceiver(nodes[source])
		nodes[target].assignCChannel(cc)
	}

	// create light source, detector and quantum channels
	for _, v := range links {
		source := int(v.(map[string]interface{})["source"].(float64))
		target := int(v.(map[string]interface{})["target"].(float64))
		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: DISTANCE, lightSpeed: LIGHTSPEED}
		qcName := fmt.Sprint("qc_", nodes[source].name, "_", nodes[target].name)
		qc := QuantumChannel{name: qcName, timeline: nodes[source].timeline, OpticalChannel: op}
		qc.init()
		lsName := fmt.Sprint(nodes[source].name, ".lightsource")
		ls := LightSource{name: lsName, timeline: nodes[source].timeline, frequency: LIGHTSOURCE_FREQUENCY, meanPhotonNum: LIGHTSOURCE_MEAN, directReceiver: &qc, wavelength: WAVELEN, encodingType: polarization()}
		ls.init()
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}, {efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}}
		qsdName := fmt.Sprint(nodes[target].name, ".qsdetector")
		qsd := QSDetector{name: qsdName, timeline: nodes[target].timeline, detectors: detectors}
		qc.setReceiver(&qsd)
		nodes[source].components["lightSource"] = &ls
		qsd._init()
		qsd.init()
		nodes[target].components["detector"] = &qsd
	}

	// create BB84
	parent_protocols := make([]*Parent, 0)
	for _, v := range links {
		source := int(v.(map[string]interface{})["source"].(float64))
		target := int(v.(map[string]interface{})["target"].(float64))
		bbName := fmt.Sprint(nodes[source].name, ".bba")
		bba := BB84{name: bbName, timeline: nodes[source].timeline, role: 0} //alice.role = 0
		bbName = fmt.Sprint(nodes[target].name, ".bbb")
		bbb := BB84{name: bbName, timeline: nodes[target].timeline, role: 1} //bob.role = 1
		bba._init()
		bbb._init()
		bba.assignNode(nodes[source], CCDELAY, int(DISTANCE/LIGHTSPEED))
		bbb.assignNode(nodes[target], CCDELAY, int(DISTANCE/LIGHTSPEED))
		bba.another = &bbb
		bbb.another = &bba
		// TODO: assign protocols to nodes
		pa := Parent{keySize: KEYSIZE, role: "alice"}
		parent_protocols = append(parent_protocols, &pa)
		pb := Parent{keySize: KEYSIZE, role: "bob"}
		pa.child = &bba
		pb.child = &bbb
		bba.addParent(&pa)
		bbb.addParent(&pb)
	}

	// schedule initial events
	for i := 0; i < len(parent_protocols); i++ {
		message := kernel.Message{}
		process := kernel.Process{Fnptr: parent_protocols[i].run, Message: message, Owner: parent_protocols[i].child.timeline}
		event := kernel.Event{Time: 0, Priority: 0, Process: &process}
		parent_protocols[i].child.timeline.Schedule(&event)
	}

	kernel.Run(tls)

	/*	for i := 0; i < totalNodes; i++ {
		fmt.Println(nodes[i].name)
		fmt.Println("   latency (s): " + fmt.Sprintf("%f", parent_protocols[i].child.latency))
		//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
		fmt.Print("   average throughput (Mb/s): ")
		fmt.Println("   ", 1e-6*floats.Sum(parent_protocols[i].child.throughPuts)/float64(len(parent_protocols[i].child.errorRates)))
		//fmt.Println("   bit error rates:")
		//for i, e := range parent_protocols[i].child.errorRates {
		//	fmt.Println("\tkey " + strconv.Itoa(i+1) + ":\t" + fmt.Sprintf("%f", e*100) + "%")
		//}
		fmt.Print("   sum error rates: ")
		fmt.Println("   ", floats.Sum(parent_protocols[i].child.errorRates)/float64(len(parent_protocols[i].child.errorRates)))
	}*/

	fmt.Println("sync counter:", tls[0].SyncCounter)
	fmt.Println("end time", tls[0].Now())
	//WriteFile(tls[0].FuncTime)
}

func randomSchedule(threadNum, nodeNum int) [][]int {
	plan := make([][]int, threadNum)
	for i := 0; i < threadNum; i++ {
		plan[i] = make([]int, 0)
	}
	for i := 0; i < nodeNum; i++ {
		thread_id := rand.Intn(threadNum)
		plan[thread_id] = append(plan[thread_id], i)
	}

	fmt.Println(plan)
	return plan
}
