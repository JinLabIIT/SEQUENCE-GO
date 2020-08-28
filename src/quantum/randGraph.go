package quantum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/matsulib/goanneal"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"kernel"
	"math"
	"os"
	"partition"
	"sync"
	"time"
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

	plan, lookAhead := randomSchedule(threadNum, graph, SIM_TIME)
	//fmt.Println(plan, lookAhead)
	if optimized {
		plan, lookAhead = optimization(graph, plan, SIM_TIME, filename, threadNum)
	}

	fmt.Println("n: ", n, "threadNum: ", threadNum, "lookAhead:", lookAhead)

	rand.Seed(SEED)
	count := 0
	photonNumber := 0
	tls := make([]*kernel.Timeline, threadNum)
	for i := 0; i < threadNum; i++ {
		tlName := fmt.Sprint("timeline", i)
		tl := kernel.Timeline{Name: tlName}
		tl.Init(uint64(lookAhead), uint64(SIM_TIME)) // 10 ms
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
		event := parent_protocols[i].child.timeline.EventPool.Get().(*kernel.Event)
		event.Time = 0
		event.Priority = 0
		event.Process.Fnptr = parent_protocols[i].run
		event.Process.Owner = parent_protocols[i].child.timeline
		parent_protocols[i].child.timeline.Schedule(event)

		//message := kernel.Message{}
		//process := kernel.Process{Fnptr: parent_protocols[i].run, Message: message, Owner: parent_protocols[i].child.timeline}
		//event := kernel.Event{Time: 0, Priority: 0, Process: &process}
		//parent_protocols[i].child.timeline.Schedule(&event)
	}

	tick := time.Now().UnixNano()
	kernel.Run(tls)
	tock := time.Now().UnixNano()
	elapsed := tock - tick
	if elapsed <= 0 {
		fmt.Println(elapsed)

	}

	var logName string

	if optimized {
		logName = "real1.log"
	} else {
		logName = "real2.log"
	}
	file, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	datawriter := bufio.NewWriter(file)
	sentence := fmt.Sprintf("%s %d %d\n", filename, threadNum, elapsed)
	datawriter.WriteString(sentence)
	datawriter.Flush()
	file.Close()

	//for i := 0; i < len(parent_protocols); i++ {
	//	fmt.Println(parent_protocols[i].child.name)
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
	//fmt.Println("end time", tls[0].Now())
	//WriteFile(tls[0].FuncTime)
	fmt.Println("No. of events created: " + fmt.Sprint(count))
	fmt.Println("No. of photons created: " + fmt.Sprint(photonNumber))
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

func randomSchedule(threadNum int, graph [][]partition.EdgeAttribute, simTime float64) ([]map[int]bool, float64) {
	plan := make([]map[int]bool, threadNum)
	for i := 0; i < threadNum; i++ {
		plan[i] = make(map[int]bool, 0)
	}
	for i := 0; i < len(graph); i++ {
		thread_id := rand.Intn(threadNum)
		plan[thread_id][i] = true
	}

	pState := partition.NewPartitionState(graph, plan, 1, simTime, 1)

	//fmt.Println(plan)
	return plan, pState.GetLookAhead()
}

func optimization(graph [][]partition.EdgeAttribute, plan []map[int]bool, simTime float64, filename string, threadNum int) ([]map[int]bool, float64) {
	pState := partition.NewPartitionState(graph, plan, 1, simTime, 1)
	randTime := pState.Energy()
	tsp := goanneal.NewAnnealer(pState)
	tsp.Steps = 100000
	afterState, energy := tsp.Anneal()
	//fmt.Println(afterState.(*partition.PartitionState).State)
	optTime := energy

	logName := "est.log"
	file, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	datawriter := bufio.NewWriter(file)
	sentence := fmt.Sprintf("%s %d %f %f\n", filename, threadNum, randTime, optTime)
	datawriter.WriteString(sentence)
	datawriter.Flush()
	file.Close()
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
