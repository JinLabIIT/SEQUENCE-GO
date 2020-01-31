package quantum

import (
	"kernel"
	"math"
	"math/rand"
)

type Detector struct {
	name              string           //inherit
	timeline          *kernel.Timeline //inherit
	efficiency        float64
	darkCount         float64
	countRate         float64
	timeResolution    uint64
	photonTimes       []uint64
	nextDetectionTime uint64
	photonCounter     int
	on                bool
}

func (d *Detector) _init() {
	if d.efficiency == 0 {
		d.efficiency = 1
	}
	if d.countRate == 0 {
		d.countRate = math.MaxFloat64 // measured in Hz
	}
	if d.timeResolution == 0 {
		d.timeResolution = 1
	}
	d.on = true
}

func (d *Detector) init() {
	d.addDarkCount(kernel.Message{})
}

func (d *Detector) get(message kernel.Message) {
	darkGet := message["darkGet"].(bool)
	d.photonCounter += 1
	now := d.timeline.Now()
	if (rand.Float64() < d.efficiency || darkGet) || (now > d.nextDetectionTime) {
		time := (now / d.timeResolution) * d.timeResolution
		d.photonTimes = append(d.photonTimes, time)
		d.nextDetectionTime = now + uint64(math.Pow10(12)/d.countRate)
	}
}

func (d *Detector) addDarkCount(message kernel.Message) {
	if d.on {
		timeToNext := uint64(rand.ExpFloat64()/d.darkCount) * uint64(math.Pow10(12))
		time := timeToNext + d.timeline.Now()
		message1 := kernel.Message{}
		process1 := kernel.Process{Fnptr: d.addDarkCount, Message: message1, Owner: d.timeline}
		event1 := kernel.Event{Time: time, Process: &process1, Priority: 0}
		message2 := kernel.Message{"darkGet": true}
		process2 := kernel.Process{Fnptr: d.get, Message: message2, Owner: d.timeline}
		event2 := kernel.Event{Time: time, Process: &process2, Priority: 0}
		d.timeline.Schedule(&event1)
		d.timeline.Schedule(&event2)
	}
}
