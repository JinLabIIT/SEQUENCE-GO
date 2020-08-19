package quantum

import (
	"kernel"
	"math"
)

type ClassicalChannel struct {
	OpticalChannel
	name     string           // inherit
	timeline *kernel.Timeline // inherit
	sender   *Node
	receiver *Node
	delay    float64 // ps
}

func (cc *ClassicalChannel) SetSender(sender *Node) {
	cc.timeline = sender.timeline
	cc.sender = sender
}

func (cc *ClassicalChannel) SetReceiver(receiver *Node) {
	cc.receiver = receiver
}

func (cc *ClassicalChannel) transmit(msg string, source *Node) {
	if cc.delay == float64(0) {
		panic("classical channel delay is 0")
	}
	futureTime := cc.timeline.Now() + uint64(math.Round(cc.delay))

	event := cc.timeline.EventPool.Get().(*kernel.Event)
	event.Time = futureTime
	event.Priority = 0
	event.Process.Message["src"] = source.name
	event.Process.Message["message"] = msg
	event.Process.Fnptr = cc.receiver.receiveMessage
	event.Process.Owner = cc.receiver.timeline
	cc.timeline.Schedule(event)
	/*
		message := kernel.Message{"src": source.name, "message": msg}
		process := kernel.Process{Fnptr: cc.receiver.receiveMessage, Message: message, Owner: cc.receiver.timeline}
		event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
		cc.timeline.Schedule(&event)
	*/
}
