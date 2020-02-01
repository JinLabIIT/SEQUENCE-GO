package quantum

import (
	"kernel"
	"math"
)

type ClassicalChannel struct {
	OpticalChannel
	name     string           // inherit
	timeline *kernel.Timeline // inherit
	ends     []*Node          // ends must equal to 2
	delay    float64
}

func (cc *ClassicalChannel) _init() {
	if cc.delay == 0 {
		cc.delay = cc.distance / cc.lightSpeed
	}
}

func (cc *ClassicalChannel) addEnd(node *Node) {
	if exists(cc.ends, node) {
		panic("already have endpoint " + node.name)
	}
	if len(cc.ends) == 2 {
		panic("channel already has 2 endpoints")
	}
	cc.ends = append(cc.ends, node)
}

func (cc *ClassicalChannel) setEnds(nodeList []*Node) {
	for _, node := range nodeList {
		cc.addEnd(node)
	}
	for _, node := range nodeList {
		node.assignCChannel(cc)
	}
}

func (cc *ClassicalChannel) transmit(msg string, source *Node) {
	if !exists(cc.ends, source) {
		panic("no endpoint " + source.name)
	}
	var receiver *Node
	for _, e := range cc.ends {
		if e != source {
			receiver = e
		}
	}
	message := kernel.Message{"message": msg}
	futureTime := cc.timeline.Now() + uint64(math.Round(cc.delay))
	process := kernel.Process{Fnptr: receiver.receiveMessage, Message: message, Owner: cc.timeline}
	event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
	cc.timeline.Schedule(&event)
}
