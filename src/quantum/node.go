package quantum

import (
	"fmt"
	"kernel"
	"math"
	"strconv"
)

type Node struct {
	name       string           // inherit
	timeline   *kernel.Timeline // inherit
	components map[string]interface{}
	count      []int
	message    kernel.Message //temporary storage for message received through classical channel
	protocol   interface{}    //
	splitter   *BeamSplitter
	receiver   *Node
}

func (node *Node) sendQubits(basisList, bitList []int, sourceName string) {
	encodingType := node.components[sourceName].(*LightSource).encodingType //???
	stateList := make(Basis, 0, len(bitList))
	for i, bit := range bitList {
		basis := (encodingType["bases"].([][][]complex128))[basisList[i]]
		state := basis[bit]
		stateList = append(stateList, state)
	}
	node.components[sourceName].(*LightSource).emit(&stateList)
}

func (node *Node) getBits(lightTime float64, startTime uint64, frequency float64, detectorName string) []int {
	encodingType := node.components[detectorName].(*QSDetector).encodingType
	length := int(math.Round(lightTime * frequency))
	bits := makeArray(length, -1) // -1 used for invalid bits
	if encodingType["name"] == "polarization" {
		// determine indices from detection times and record bits
		detectionTimes := node.components[detectorName].(*QSDetector).getPhotonTimes()
		for _, time := range detectionTimes[0] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				bits[index] = 0
			}
		}
		for _, time := range detectionTimes[1] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				if bits[index] == 0 {
					bits[index] = -1
				} else {
					bits[index] = 1
				}
			}
		}
		return bits
	} else if encodingType["name"] == "timeBin" {
		detectionTimes := node.components[detectorName].(*QSDetector).getPhotonTimes()
		binSeparation, _ := encodingType["binSeparation"].(float64)
		// single detector (for early, late basis) times
		for _, time := range detectionTimes[0] {
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if 0 <= index && index < len(bits) {
				if math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
					bits[index] = 0
				} else if math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime)-(float64(time)-binSeparation)) < binSeparation/2 {
					bits[index] = 1
				}
			}
		}
		for _, time := range detectionTimes[1] {
			time -= uint64(binSeparation)
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if (0 <= index && index < len(bits)) && math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
				if bits[index] == -1 {
					bits[index] = 0
				} else {
					bits[index] = -1
				}
			}
		}
		for _, time := range detectionTimes[2] {
			time -= uint64(binSeparation)
			index := int(math.Round(float64(time-startTime) * frequency * math.Pow10(-12)))
			if (0 <= index && index < len(bits)) && math.Abs(float64(index)*math.Pow10(12)/frequency+float64(startTime-time)) < binSeparation/2 {
				if bits[index] == -1 {
					bits[index] = 1
				} else {
					bits[index] = -1
				}
			}
		}
		return bits
	} else {
		panic("Invalid encoding type for node " + node.name)
	}
}

func (node *Node) setBases(basisList []int, startTime uint64, frequency float64, detectorName string) {
	encodingType := node.components[detectorName].(*QSDetector).encodingType
	basisStartTime := startTime - uint64(math.Pow10(12)/(2*frequency))
	if encodingType["name"] == "polarization" {
		splitter := node.components[detectorName].(*QSDetector).splitter
		splitter.startTime = basisStartTime
		splitter.frequency = frequency

		splitterBasisList := make([]*Basis, 0, len(basisList))
		for _, d := range basisList {
			base := encodingType["bases"].([][][]complex128)
			tmp := Basis{base[d][0], base[d][1]}
			splitterBasisList = append(splitterBasisList, &tmp)
		}
		splitter.basisList = splitterBasisList
	} else if encodingType["name"] == "timeBin" {
		_switch := node.components[detectorName].(*QSDetector)._switch
		_switch.startTime = basisStartTime
		_switch.frequency = frequency
		_switch.stateList = basisList
	} else {
		panic("Invalid encoding type for node " + node.name)
	}
}

func (node *Node) getSourceCount() interface{} { // tmp
	source := node.components["lightSource"]
	return source
}

func (node *Node) sendMessage(msg string, channel string) {
	fmt.Println("sendMessage " + strconv.FormatUint(node.timeline.Now(), 10))
	node.components[channel].(*ClassicalChannel).transmit(msg, node.receiver)
}

func (node *Node) receiveMessage(message kernel.Message) {
	node.message = message
	node.protocol.(*BB84).receivedMessage()
}
