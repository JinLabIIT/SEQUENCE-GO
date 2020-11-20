package quantum

import (
	"encoding/json"
	"fmt"
	_ "github.com/gonum/floats"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"kernel"
	"os"
	"strconv"
	"time"
)

func RandGraphNN(path string, base_seed int) {
	fmt.Println(path, base_seed)
	SEED := uint64(base_seed)
	rand.Seed(SEED)

	SIM_TIME := uint64(rand.Float32()*2e10 + 1e10)

	ATTENUATION := 0.0002
	QCFIDELITY := 0.99
	LIGHTSPEED := 2e-4
	LIGHTSOURCE_MEAN := 0.1
	CCDELAY := 1e9

	WAVELEN := 1550.0
	DETECTOR_EFFICIENCY := 0.8
	DARKCOUNT := 0.0
	TIME_RESOLUTION := uint64(10)
	COUNT_RATE := 5e7

	KEYSIZE := 512

	jsonFile, err := os.Open(path + "/graph.json")
	if err != nil {
		fmt.Println(err)
	}
	var result map[string]interface{}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	//byteValue = byteValue
	err = json.Unmarshal(byteValue, &result)

	nodesGraph := result["nodes"].([]interface{})
	n := len(nodesGraph)
	links := result["links"].([]interface{}) // links: {Source: int,Target: int}

	randomLinks := randomAttributes(links)
	graph := createGraph(n, randomLinks, LIGHTSOURCE_MEAN, ATTENUATION, LIGHTSPEED)
	threadNum := rand.Intn(32) + 1

	plan, lookAhead := randomSchedule(threadNum, graph)

	//json_plan, _ := json.Marshal(plan)
	//fmt.Println(string(json_plan))
	//
	//json_content, _ := json.Marshal(randomLinks)
	//fmt.Println(string(json_content))

	//fmt.Println("n: ", n, "threadNum: ", threadNum, "lookAhead:", lookAhead)

	tls := make([]*kernel.Timeline, threadNum)
	for i := 0; i < threadNum; i++ {
		tlName := fmt.Sprint("timeline", i)
		tl := kernel.Timeline{Name: tlName}
		tl.Init(uint64(lookAhead), uint64(SIM_TIME)) // 10 ms
		tls[i] = &tl
	}

	// create nodes

	totalNodes := n
	nodes := make([]*Node, totalNodes)
	for threadId := range plan {
		for nodeId := range plan[threadId] {
			nodeName := fmt.Sprint("node", nodeId)
			node := Node{name: nodeName, timeline: tls[threadId]}
			node.cchannels = make(map[string]*ClassicalChannel)
			node.components = make(map[string]interface{})
			nodes[nodeId] = &node
		}
	}

	// create classical channels

	for _, link := range randomLinks {
		source := link.Source
		target := link.Target

		if nodes[source].HasCCto(nodes[target]) {
			continue
		}

		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: link.Distance, lightSpeed: LIGHTSPEED}
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

	// create light Source, detector and quantum channels
	for i, link := range randomLinks {
		source := link.Source
		target := link.Target
		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: link.Distance, lightSpeed: LIGHTSPEED}
		qcName := fmt.Sprint("qc_", nodes[source].name, "_", nodes[target].name)
		qc := QuantumChannel{name: qcName, timeline: nodes[source].timeline, OpticalChannel: op}
		qc.init()
		lsName := fmt.Sprint(nodes[source].name, ".lightsource")
		ls := LightSource{name: lsName, timeline: nodes[source].timeline, frequency: link.Frequency, meanPhotonNum: LIGHTSOURCE_MEAN, directReceiver: &qc, wavelength: WAVELEN, encodingType: polarization()}
		ls.init(int64(i))
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}, {efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}}
		qsdName := fmt.Sprint(nodes[target].name, ".qsdetector")
		qsd := QSDetector{name: qsdName, timeline: nodes[target].timeline, detectors: detectors}
		qc.setReceiver(&qsd)
		sourceName := fmt.Sprintf("lightSource.%d.%d", source, target)
		nodes[source].components[sourceName] = &ls
		qsd._init()
		qsd.init(int64(i))
		detectorName := fmt.Sprintf("detector.%d.%d", source, target)
		nodes[target].components[detectorName] = &qsd
	}

	// create BB84
	parent_protocols := make([]*Parent, 0)
	for i, link := range randomLinks {
		source := link.Source
		target := link.Target
		sourceName := fmt.Sprintf("lightSource.%d.%d", source, target)
		detectorName := fmt.Sprintf("detector.%d.%d", source, target)
		bbName := fmt.Sprint(nodes[source].name, ".bba.", target)
		bba := BB84{name: bbName, timeline: nodes[source].timeline, role: 0, sourceName: sourceName, detectorName: detectorName} //alice.role = 0
		bbName = fmt.Sprint(nodes[target].name, ".bbb.", target)
		bbb := BB84{name: bbName, timeline: nodes[target].timeline, role: 1, sourceName: sourceName, detectorName: detectorName} //bob.role = 1
		bba._init(int64(i * 2))
		bbb._init(int64(i*2 + 1))
		bba.assignNode(nodes[source], CCDELAY, int(link.Distance/LIGHTSPEED))
		bbb.assignNode(nodes[target], CCDELAY, int(link.Distance/LIGHTSPEED))
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

	tick := time.Now()
	kernel.Run(tls)
	tock := time.Now()

	all_content := map[string]interface{}{}

	all_content["node_number"] = n
	all_content["thread_number"] = threadNum
	all_content["sim_time"] = SIM_TIME
	all_content["exe_time"] = tock.Sub(tick)
	all_content["plan"] = plan
	all_content["links"] = randomLinks
	json_all_content, _ := json.Marshal(all_content)
	log_filename := path + "/log" + strconv.Itoa(base_seed) + ".json"
	fmt.Println(log_filename)
	_ = ioutil.WriteFile(log_filename, json_all_content, 0644)
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

	//fmt.Println("sync counter:", tls[0].SyncCounter)
	//fmt.Println("end time", tls[0].Now())
	//WriteFile(tls[0].FuncTime)
	//name := "ExceTime.txt"
	//for _ , tl := range tls{
	//	WriteUint64(tl.SyncWindowsEvent,tl.SyncWindowsTime,threadNum,name)
	//}
}
