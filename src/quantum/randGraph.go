package quantum

import (
	"encoding/json"
	"fmt"
	_ "github.com/gonum/floats"
	"github.com/matsulib/goanneal"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"kernel"
	"math"
	"os"
	"partition"
)

func RandGraph(threadNum int, filename string, optimized bool) {
	SEED := uint64(0)
	SIM_TIME := 2e10

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

	jsonFile, err := os.Open(filename)
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

	randomLinks := randomAttributes(links)
	graph := createGraph(n, randomLinks, LIGHTSOURCE_MEAN, ATTENUATION, LIGHTSPEED)

	plan, lookAhead := randomSchedule(threadNum, graph)
	if optimized {
		plan, lookAhead = optimization(graph, plan)
	}

	fmt.Println("n: ", n, "threadNum: ", threadNum, "lookAhead:", lookAhead)

	rand.Seed(SEED)

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
		source := link.source
		target := link.target

		if nodes[source].HasCCto(nodes[target]) {
			continue
		}

		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: link.distance, lightSpeed: LIGHTSPEED}
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
	for _, link := range randomLinks {
		source := link.source
		target := link.target
		op := OpticalChannel{polarizationFidelity: QCFIDELITY, attenuation: ATTENUATION, distance: link.distance, lightSpeed: LIGHTSPEED}
		qcName := fmt.Sprint("qc_", nodes[source].name, "_", nodes[target].name)
		qc := QuantumChannel{name: qcName, timeline: nodes[source].timeline, OpticalChannel: op}
		qc.init()
		lsName := fmt.Sprint(nodes[source].name, ".lightsource")
		ls := LightSource{name: lsName, timeline: nodes[source].timeline, frequency: link.frequency, meanPhotonNum: LIGHTSOURCE_MEAN, directReceiver: &qc, wavelength: WAVELEN, encodingType: polarization()}
		ls.init()
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}, {efficiency: DETECTOR_EFFICIENCY, darkCount: DARKCOUNT, timeResolution: TIME_RESOLUTION, countRate: COUNT_RATE}}
		qsdName := fmt.Sprint(nodes[target].name, ".qsdetector")
		qsd := QSDetector{name: qsdName, timeline: nodes[target].timeline, detectors: detectors}
		qc.setReceiver(&qsd)
		sourceName := fmt.Sprintf("lightSource.%d.%d", source, target)
		nodes[source].components[sourceName] = &ls
		qsd._init()
		qsd.init()
		detectorName := fmt.Sprintf("detector.%d.%d", source, target)
		nodes[target].components[detectorName] = &qsd
	}

	// create BB84
	parent_protocols := make([]*Parent, 0)
	for _, link := range randomLinks {
		source := link.source
		target := link.target
		sourceName := fmt.Sprintf("lightSource.%d.%d", source, target)
		detectorName := fmt.Sprintf("detector.%d.%d", source, target)
		bbName := fmt.Sprint(nodes[source].name, ".bba.", target)
		bba := BB84{name: bbName, timeline: nodes[source].timeline, role: 0, sourceName: sourceName, detectorName: detectorName} //alice.role = 0
		bbName = fmt.Sprint(nodes[target].name, ".bbb.", target)
		bbb := BB84{name: bbName, timeline: nodes[target].timeline, role: 1, sourceName: sourceName, detectorName: detectorName} //bob.role = 1
		bba._init()
		bbb._init()
		bba.assignNode(nodes[source], CCDELAY, int(link.distance/LIGHTSPEED))
		bbb.assignNode(nodes[target], CCDELAY, int(link.distance/LIGHTSPEED))
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

	//fmt.Println("sync counter:", tls[0].SyncCounter)
	//fmt.Println("end time", tls[0].Now())
	//WriteFile(tls[0].FuncTime)
	//name := "ExceTime.txt"
	//for _ , tl := range tls{
	//	WriteUint64(tl.SyncWindowsEvent,tl.SyncWindowsTime,threadNum,name)
	//}
}

type Link struct {
	source    int
	target    int
	frequency float64
	distance  float64
}

func randomAttributes(links []interface{}) []*Link {
	FREQ_UPPER := 1e8
	FREQ_LOWER := 1e6
	DIST_UPPER := 15e3
	DIST_LOWER := 5e3
	res := []*Link{}
	for _, link := range links {
		source := int(link.(map[string]interface{})["source"].(float64))
		target := int(link.(map[string]interface{})["target"].(float64))
		frequency := float64(rand.Intn(int(FREQ_UPPER-FREQ_LOWER))) + FREQ_LOWER
		distance := float64(rand.Intn(int(DIST_UPPER-DIST_LOWER))) + DIST_LOWER
		res = append(res, &Link{
			source:    source,
			target:    target,
			frequency: frequency,
			distance:  distance,
		})
	}
	return res
}

func randomSchedule(threadNum int, graph [][]partition.EdgeAttribute) ([]map[int]bool, float64) {
	plan := make([]map[int]bool, threadNum)
	for i := 0; i < threadNum; i++ {
		plan[i] = make(map[int]bool, 0)
	}
	for i := 0; i < len(graph); i++ {
		thread_id := rand.Intn(threadNum)
		plan[thread_id][i] = true
	}

	pState := partition.NewPartitionState(graph, plan, 1, 1)

	//fmt.Println(plan)
	return plan, pState.GetLookAhead()
}

func optimization(graph [][]partition.EdgeAttribute, plan []map[int]bool) ([]map[int]bool, float64) {
	pState := partition.NewPartitionState(graph, plan, 1, 1)
	fmt.Println("initial plan: estimated exe time ", pState.Energy()/1e9, "sec")
	tsp := goanneal.NewAnnealer(pState)
	tsp.Steps = 100000
	afterState, energy := tsp.Anneal()
	//fmt.Println(afterState.(*partition.PartitionState).State)
	fmt.Println("optimized plan: estimated exe time ", energy/1e9, "sec")
	return afterState.(*partition.PartitionState).State, afterState.(*partition.PartitionState).GetLookAhead()
}

func createGraph(nodeNum int, links []*Link, mean, attenuation, lightspeed float64) [][]partition.EdgeAttribute {
	graph := make([][]partition.EdgeAttribute, nodeNum)
	for i := 0; i < nodeNum; i++ {
		graph[i] = make([]partition.EdgeAttribute, nodeNum)
	}

	for _, link := range links {
		source := link.source
		target := link.target
		weight := link.frequency * 1e-12 * mean
		ratio := math.Pow(10, link.distance*attenuation/-10)
		qcdelay := link.distance / lightspeed
		graph[source][target] = partition.EdgeAttribute{
			Weight:    weight,
			Ratio:     ratio,
			LookAhead: int64(qcdelay),
		}
	}
	return graph
}
