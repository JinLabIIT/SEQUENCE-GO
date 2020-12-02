package quantum

import (
	"kernel"
	"math"
)

type Node struct {
	name               string           // inherit
	timeline           *kernel.Timeline // inherit
	components         map[string]interface{}
	count              []int
	protocols          []interface{} //
	splitter           *BeamSplitter
	receiver           *Node
	cchannels          map[string]*ClassicalChannel
	cc_message_counter int
}

func (node *Node) HasCCto(dst *Node) bool {
	_, exist := node.cchannels[dst.name]
	return exist
}

func (node *Node) assignCChannel(cc *ClassicalChannel) {
	_, exists := node.cchannels[cc.receiver.name]
	if exists {
		panic("duplicated classical channel is assigned")
	}
	node.cchannels[cc.receiver.name] = cc
}

func (node *Node) sendQubits(basisList, bitList []int, sourceName string) {
	encodingType := node.components[sourceName].(*LightSource).encodingType //???
	stateList := make([][]complex128, 0, len(bitList))
	for i, bit := range bitList {
		basis := (encodingType["bases"].([][2][2]complex128))[basisList[i]]
		//state := basis[bit]
		state := []complex128{basis[bit][0], basis[bit][1]}
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

		splitterBasisList := make([][2][2]complex128, 0, len(basisList))
		for _, d := range basisList {
			base := encodingType["bases"].([][2][2]complex128)
			tmp := [2][2]complex128{base[d][0], base[d][1]}
			splitterBasisList = append(splitterBasisList, tmp)
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

func (node *Node) sendMessage(msg string, dst string) {
	//node.components[channel].(*ClassicalChannel).transmit(msg, node.receiver)
	node.cchannels[dst].transmit(msg, node)
}

func (node *Node) receiveMessage(message kernel.Message) {
	node.cc_message_counter++
	for _, protocol := range node.protocols {
		protocol.(*BB84).receivedMessage(message)
	}
}
