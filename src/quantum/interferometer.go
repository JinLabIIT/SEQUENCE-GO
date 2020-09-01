package quantum

import (
	"kernel"
	"math"
	"math/rand"
)

type Interferometer struct {
	timeline       *kernel.Timeline // inherit
	pathDifference int
	phaseError     float64
	detectors      []*Detector
}

func (inf *Interferometer) get(photon *Photon) {
	detectorNum := rand.Intn(2)
	quantumState := []complex128{photon.quantumState[0], photon.quantumState[1]}
	time := 0
	random := rand.Float64()

	if quantumState[0] == complex(1, 0) && quantumState[1] == complex(0, 0) { // early
		if random <= 0.5 {
			time = 0
		} else {
			time = inf.pathDifference
		}
	}
	if quantumState[0] == complex(0, 0) && quantumState[1] == complex(1, 0) { // late
		if random <= 0.5 {
			time = inf.pathDifference
		} else {
			time = 2 * inf.pathDifference
		}
	}
	if rand.Float64() < inf.phaseError {
		quantumState = multiply([]float64{1, -1}, quantumState) // list??
	}
	if quantumState[0] == complex(math.Sqrt(0.5), 0) && quantumState[1] == complex(math.Sqrt(0.5), 0) { // early + late
		if random <= 0.25 {
			time = 0
		} else if random <= 0.5 {
			time = 2 * inf.pathDifference
		} else if detectorNum == 0 {
			time = inf.pathDifference
		} else {
			return
		}
		if quantumState[0] == complex(math.Sqrt(0.5), 0) && quantumState[1] == complex(math.Sqrt(-0.5), 0) { // early - late
			if random <= 0.25 {
				time = 0
			} else if random <= 0.5 {
				time = 2 * inf.pathDifference
			} else if detectorNum == 1 {
				time = inf.pathDifference
			} else {
				return
			}
		}
		message := kernel.Message{}
		process := kernel.Process{Fnptr: inf.detectors[detectorNum].get, Message: message, Owner: inf.timeline}
		event := kernel.Event{Time: inf.timeline.Now() + uint64(time), Process: &process, Priority: 0}
		inf.timeline.Schedule(&event)
	}
}

type Switch struct {
	timeline  *kernel.Timeline
	receiver  []interface{} // Interferometer Detector
	startTime uint64
	frequency float64
	stateList []int // tmp
	typeList  []int //0: Interferometer 1: Detector ???
}

func (_switch *Switch) addReceiver(entity *interface{}) {
	_switch.receiver = append(_switch.receiver, entity)
}

func (_switch *Switch) setState(state int) {
	_switch.stateList = []int{state}
}

func (_switch *Switch) get(photon *Photon) {
	index := int(float64(_switch.timeline.Now()-_switch.startTime) * _switch.frequency * math.Pow10(-12))
	if index < 0 || index >= len(_switch.stateList) {
		index = 0
	}
	receiver := _switch.receiver[_switch.stateList[index]]
	// check if receiver is detector, if we're using time bin, and if the photon is "late" to schedule measurement
	if _switch.typeList[index] == 1 { //???
		if photon.encodingType["name"] == "timeBin" && photon.measure(photon.encodingType["bases"].([]*[2][2]complex128)[0], 0.0) == 1 { //question mark
			time := _switch.timeline.Now() + photon.encodingType["binSeparation"].(uint64)
			message := kernel.Message{}
			process := kernel.Process{Fnptr: receiver.(*Detector).get, Message: message, Owner: _switch.timeline}
			event := kernel.Event{Time: time, Priority: 0, Process: &process}
			_switch.timeline.Schedule(&event)
		} else {
			receiver.(*Detector).get(kernel.Message{})
		}
	} else {
		receiver.(*Interferometer).get(photon)
	}

}
