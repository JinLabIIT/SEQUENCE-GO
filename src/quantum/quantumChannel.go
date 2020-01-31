package quantum

import (
	"kernel"
	"math"
	"math/rand"
)

type QuantumChannel struct {
	OpticalChannel
	name          string           // inherit
	timeline      *kernel.Timeline // inherit
	sender        *LightSource
	receiver      *QSDetector //tmp
	depoCount     int
	photonCounter int
}

func (qc *QuantumChannel) setSender(sender *LightSource) {
	qc.sender = sender
}

func (qc *QuantumChannel) setReceiver(receiver *QSDetector) {
	qc.receiver = receiver
}

func (qc *QuantumChannel) get(photon *Photon) {
	loss := qc.distance * qc.attenuation
	chancePhotonKept := math.Pow(10, loss/-10)
	// check if photon kept
	if rand.Float64() < chancePhotonKept { // numpy.random.random_sample()
		qc.photonCounter += 1
		if rand.Float64() > qc.polarizationFidelity && photon.encodingType["name"] == "polarization" {
			photon.randomNoise()
			qc.depoCount += 1
		}
		futureTime := qc.timeline.Now() + uint64(math.Round(qc.distance/qc.lightSpeed))
		message := kernel.Message{"photon": photon}
		process := kernel.Process{Fnptr: qc.receiver.get, Message: message, Owner: qc.timeline}
		event := kernel.Event{Time: futureTime, Process: &process, Priority: 0}
		qc.timeline.Schedule(&event)
	}
}
