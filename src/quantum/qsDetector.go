package quantum

import (
	"kernel"
	"reflect"
)

type QSDetector struct {
	name           string           // inherit
	timeline       *kernel.Timeline // inherit
	encodingType   map[string]interface{}
	detectors      []*Detector // tmp
	splitter       *BeamSplitter
	_switch        *Switch
	interferometer *Interferometer
	eventCounter   int
	first_get_time uint64
}

func (qsd *QSDetector) _init() {
	if qsd.encodingType == nil {
		qsd.encodingType = polarization()
	}
	//fmt.Println(qsd.encodingType["name"])
	if (qsd.encodingType["name"] == "polarization" && len(qsd.detectors) != 2) ||
		(qsd.encodingType["name"] == "timeBin" && len(qsd.detectors) != 3) {
		panic("invalid number of detectors specified")
	}
	for i := range qsd.detectors {
		if !reflect.DeepEqual(qsd.detectors[i], Detector{}) { // question mark
			qsd.detectors[i].timeline = qsd.timeline
		} else {
			qsd.detectors[i] = &Detector{}
		}
	}
	if qsd.encodingType["name"] == "polarization" {
		bs := BeamSplitter{timeline: qsd.timeline}
		bs._init()
		qsd.splitter = &bs
	} else if qsd.encodingType["name"] == "timeBin" {
		qsd.interferometer = &Interferometer{timeline: qsd.timeline}
		qsd.interferometer.detectors = qsd.detectors[1:]
		qsd._switch = &Switch{timeline: qsd.timeline}
		qsd._switch.receiver = make([]interface{}, 0)
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.detectors[0])
		qsd._switch.receiver = append(qsd._switch.receiver, qsd.interferometer)
		qsd._switch.typeList = []int{1, 0}
	} else {
		panic("invalid encoding type for QSDetector " + qsd.name)
	}
}

func (qsd *QSDetector) init(seed int64) {
	for _, d := range qsd.detectors {
		if !reflect.DeepEqual(d, Detector{}) {
			d.init(seed)
		}
	}
}

func (qsd *QSDetector) get(message kernel.Message) {
	photon := message["photon"].(*Photon)
	if qsd.first_get_time == 0 {
		qsd.first_get_time = qsd.timeline.Now()
		//fmt.Println(qsd.first_get_time)
	}
	qsd.eventCounter += 1
	//if qsd.timeline.Now() >= qsd.timeline.OutputTime {
	//	fmt.Println(qsd.timeline.Now(), qsd.eventCounter)
	//}
	if qsd.encodingType["name"].(string) == "polarization" {
		detector := qsd.splitter.get(photon)
		//if detector == 0 || detector == 1 {
		//detector = qsd.splitter.get(photon)//test
		qsd.detectors[detector].get(kernel.Message{"darkGet": false}) //??????
		//}
	} else if qsd.encodingType["name"].(string) == "timeBin" {
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

func (qsd *QSDetector) setBasis(basis *[2][2]complex128) {
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
			//d.init()
			d.on = true
		}
	}
}
