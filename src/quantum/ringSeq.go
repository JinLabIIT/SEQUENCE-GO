package quantum

import (
	"github.com/gonum/floats"
	"github.com/leesper/go_rng"
	"golang.org/x/exp/errors/fmt"
	"kernel"
	"math"
	"math/rand"
	"strconv"
)

func main(n int) {
	fmt.Println("Ring QKD Network (sequential)")

	seed := int64(156)
	rand.Seed(seed)
	poisson := rng.NewPoissonGenerator(seed)
	tl := kernel.Timeline{Name: "timeline", LookAhead: math.MaxInt64}
	tl.SetEndTime(uint64(math.Pow10(11))) //stop time is 100 ms

	// create nodes
	totalNodes := n
	nodes := make([]*Node, totalNodes)

	for i := 0; i < totalNodes; i++ {
		node_name := fmt.Sprint("node", i)

		node := Node{name: node_name, timeline: &tl}
		node.cchannels = make(map[string]*ClassicalChannel)
		node.components = make(map[string]interface{})
		nodes[i] = &node
	}

	// create classical channels

	for i := 0; i < totalNodes; i++ {
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		cc_name := fmt.Sprint("cc_", nodes[i].name, "_", nodes[(i+1)%totalNodes].name)
		cc := &ClassicalChannel{name: cc_name, timeline: &tl, OpticalChannel: op, delay: float64(1 * math.Pow10(9))}
		cc.setEnds([]*Node{nodes[i], nodes[(i+1)%totalNodes]})
	}

	// create light source, detector and quantum channels
	for i := 0; i < totalNodes; i++ {
		op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3), lightSpeed: 2 * math.Pow10(-4)}
		qc_name := fmt.Sprint("qc_", nodes[i].name, "_", nodes[(i+1)%totalNodes].name)
		qc := QuantumChannel{name: qc_name, timeline: &tl, OpticalChannel: op}
		ls_name := fmt.Sprint(nodes[i].name, ".lightsource")
		ls := LightSource{name: ls_name, timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc, poisson: poisson, wavelength: 1550, encodingType: polarization()}
		qc.setSender(&ls)
		detectors := []*Detector{{efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}, {efficiency: 0.8, darkCount: 0, timeResolution: 10, countRate: 50 * math.Pow10(6)}}
		qsd_name := fmt.Sprint(nodes[(i+1)%totalNodes].name, ".qsdetector")
		qsd := QSDetector{name: qsd_name, timeline: &tl, detectors: detectors}
		qc.setReceiver(&qsd)
		nodes[i].components["lightSource"] = &ls
		qsd._init()
		qsd.init()
		nodes[(i+1)%totalNodes].components["detector"] = &qsd
	}

	// create BB84
	parent_protocols := make([]*Parent, 0)
	for i := 0; i < totalNodes; i++ {
		bb_name := fmt.Sprint(nodes[i].name, ".bba")
		bba := BB84{name: bb_name, timeline: &tl, role: 0} //alice.role = 0
		bb_name = fmt.Sprint(nodes[(i+1)%totalNodes].name, ".bbb")
		bbb := BB84{name: bb_name, timeline: &tl, role: 1} //bob.role = 1
		bba._init()
		bbb._init()
		bba.assignNode(nodes[i], float64(1*math.Pow10(9)), 50000000)
		bbb.assignNode(nodes[(i+1)%totalNodes], float64(1*math.Pow10(9)), 50000000)
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
		process := kernel.Process{Fnptr: parent_protocols[i].run, Message: message, Owner: &tl}
		event := kernel.Event{Time: 0, Priority: 0, Process: &process}
		tl.Schedule(&event)
	}

	kernel.Run([]*kernel.Timeline{&tl})

	for i := 0; i < totalNodes; i++ {
		fmt.Println("latency (s): " + fmt.Sprintf("%f", parent_protocols[i].child.latency))
		//fmt.Println("average throughput (Mb/s): "+fmt.Sprintf("%f",math.Pow10(-6) * sum(bba.throughputs) / len(bba.throughputs)))
		fmt.Print("average throughput (Mb/s): ")
		fmt.Println(1e-6 * floats.Sum(parent_protocols[i].child.throughPuts) / float64(len(parent_protocols[i].child.errorRates)))
		fmt.Println("bit error rates:")
		for i, e := range parent_protocols[i].child.errorRates {
			fmt.Println("\tkey " + strconv.Itoa(i+1) + ":\t" + fmt.Sprintf("%f", e*100) + "%")
		}
		fmt.Println("sum error rates: ")
		fmt.Print(floats.Sum(parent_protocols[i].child.errorRates) / float64(len(parent_protocols[i].child.errorRates)))
	}
}
