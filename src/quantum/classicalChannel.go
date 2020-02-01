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
	delay    float64          // ps
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

func (cc *ClassicalChannel) transmit(msg string, source *Node) {
	if !exists(cc.ends, source) {
		panic("no endpoint " + source.name)
	}

	if cc.delay == float64(0) {
		panic("classical channel delay is 0")
	}
	println("transmit")

	message := kernel.Message{"message": msg}
	futureTime := cc.timeline.Now() + uint64(math.Round(cc.delay))
	process := kernel.Process{Fnptr: source.receiveMessage, Message: message, Owner: cc.timeline}
	event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
	cc.timeline.Schedule(&event)
}
