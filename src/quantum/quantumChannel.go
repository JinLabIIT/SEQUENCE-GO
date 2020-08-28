package quantum

import (
	rng "github.com/leesper/go_rng"
	"kernel"
	"math"
)

type QuantumChannel struct {
	OpticalChannel
	name          string           // inherit
	timeline      *kernel.Timeline // inherit
	sender        *LightSource
	receiver      *QSDetector //tmp
	depoCount     int
	photonCounter int
	rng           *rng.UniformGenerator
}

func (qc *QuantumChannel) init() {
	qc.rng = rng.NewUniformGenerator(123)
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
	if qc.rng.Float64() < chancePhotonKept { // numpy.random.random_sample()
		qc.photonCounter += 1
		if qc.rng.Float64() > qc.polarizationFidelity && photon.encodingType["name"] == "polarization" {
			photon.randomNoise(qc.rng.Float64())
			qc.depoCount += 1
		}
		futureTime := qc.timeline.Now() + uint64(math.Round(qc.distance/qc.lightSpeed))

		event := qc.timeline.EventPool.Get().(*kernel.Event)
		event.Time = futureTime
		event.Priority = 0
		event.Process.Message["photon"] = photon
		event.Process.Fnptr = qc.receiver.get
		event.Process.Owner = qc.receiver.timeline
		qc.timeline.Schedule(event)

	} else {
		qc.timeline.PhotonPool.Put(photon)
	}
}
