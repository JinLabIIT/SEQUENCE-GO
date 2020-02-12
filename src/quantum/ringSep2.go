package quantum

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"kernel"
	"math"
	"os"
)

func ringSep2(threadNum int, lookAhead uint64, path string) {
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
	tls := make([]*kernel.Timeline, threadNum)
	for i := 0; i < threadNum; i++ {
		tlName := fmt.Sprint("timeline", i)
		tl := kernel.Timeline{Name: tlName}
		tl.Init(lookAhead, uint64(math.Pow10(10))) // 10 ms
		tls[i] = &tl
	}

	// create nodes
	totalNodes := n
	nodes := make([]*Node, totalNodes)

	for i := 0; i < totalNodes; i++ {
		nodeName := fmt.Sprint("node", i)
		node := Node{name: nodeName, timeline: tls[rand.Intn(threadNum)]}
		node.cchannels = make(map[string]*ClassicalChannel)
		node.components = make(map[string]interface{})
		nodes[i] = &node
	}

	// create classical channels

	for _, v := range links {
		source := int(v.(map[string]interface{})["source"].(float64))
		target := int(v.(map[string]interface{})["target"].(float64))
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		ccName := fmt.Sprint("cc_", nodes[source].name, "_", nodes[target].name)
		cc := &ClassicalChannel{name: ccName, OpticalChannel: op, delay: 1 * math.Pow10(9)}
		cc.SetSender(nodes[source])
		cc.SetReceiver(nodes[target])
		nodes[source].assignCChannel(cc)

		ccName = fmt.Sprint("cc_", nodes[target].name, "_", nodes[source].name)
		cc = &ClassicalChannel{name: ccName, OpticalChannel: op, delay: 1 * math.Pow10(9)}
		cc.SetSender(nodes[target])
		cc.SetReceiver(nodes[source])
		nodes[target].assignCChannel(cc)
	}

	// create light source, detector and quantum channels
	for _, v := range links {
		source := int(v.(map[string]interface{})["source"].(float64))
		target := int(v.(map[string]interface{})["target"].(float64))
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		qcName := fmt.Sprint("qc_", nodes[source].name, "_", nodes[target].name)
		qc := QuantumChannel{name: qcName, timeline: nodes[source].timeline, OpticalChannel: op}
		qc.init()
		lsName := fmt.Sprint(nodes[source].name, ".lightsource")
		ls := LightSource{name: lsName, timeline: nodes[source].timeline, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, wavelength: 1550, encodingType: polarization()}
		ls.init()
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}}
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
		bba.assignNode(nodes[source], 1*math.Pow10(9), 50000000)
		bbb.assignNode(nodes[target], 1*math.Pow10(9), 50000000)
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
