package quantum

import (
	"testing"
)

func Test100(t *testing.T) {
	/*	var wg sync.WaitGroup
		for i := 0; i < 2; i++{
			wg.Add(1)
			go createRandom(2e7,&wg)
		}
		wg.Wait()*/
}

/*func Test3(t *testing.T) {
	//defer profile.Start().Stop()
	seed := int64(60616)
	rand.Seed(seed)
	lookahead := uint64(math.Pow10(2))
	//flag := 1 //
	for i:=10 ; i<15;i++{
		lookahead = uint64(500*i+100)
		tmp := make([]float64,2)
		for j:=0; j<2;j++{
			flag := j
			if (flag == 0){
				start := time.Now()
				tl := kernel.Timeline{Name: "alice", LookAhead: lookahead}
				endTime := uint64(8.1 * math.Pow10(4))
				tl.Init(lookahead,endTime)

				op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3)}
				op._init()
				qc1 := qc{name: "qc", timeline: &tl, OpticalChannel: op}
				qc1._init()
				poisson1 := rng.NewPoissonGenerator(seed)
				ls1 := ls{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc1, poisson: poisson1}
				ls1._init()
				ls1.bs = MakeStateList()
				qsd1 := qsd{name: "bob.qsdetector", timeline: &tl}
				qc1.receiver = &qsd1

				op2 := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3)}
				op2._init()
				qc2 := qc{name: "qc", timeline: &tl, OpticalChannel: op}
				qc2._init()
				poisson2 := rng.NewPoissonGenerator(seed)
				ls2 := ls{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc2, poisson: poisson2}
				ls2._init()
				ls2.bs = MakeStateList()
				qsd2 := qsd{name: "bob.qsdetector", timeline: &tl}
				qc2.receiver = &qsd2

				message := kernel.Message{}
				process := kernel.Process{
					Fnptr:   ls1.transmit,
					Message: message,
					Owner:   ls1.timeline,
				}
				event := kernel.Event{
					Time:     0,
					Priority: 0,
					Process:  &process,
				}

				message2 := kernel.Message{}
				process2 := kernel.Process{
					Fnptr:   ls2.transmit,
					Message: message2,
					Owner:   ls2.timeline,
				}
				event2 := kernel.Event{
					Time:     0,
					Priority: 0,
					Process:  &process2,
				}

				tl.Schedule(&event)
				tl.Schedule(&event2)
				tl.SetEndTime(uint64(8.1 * math.Pow10(4)))
				tl.SetEndTime(uint64(8.1 * math.Pow10(4)))

				kernel.Run([]*kernel.Timeline{&tl})
				//fmt.Println(ls1.tran)
				t := time.Now()
				elapsed := t.Sub(start)
				fmt.Println(elapsed)
				tmp[0] = float64(elapsed)
			}else {
				start := time.Now()
				tl := kernel.Timeline{Name: "alice", LookAhead: lookahead}
				tl2 := kernel.Timeline{Name: "bob", LookAhead: lookahead}
				endTime := uint64(8.1 * math.Pow10(4))
				tl.Init(lookahead,endTime)
				tl2.Init(lookahead,endTime)
				//tl2 = tl
				op := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3)}
				op._init()
				qc1 := qc{name: "qc", timeline: &tl2, OpticalChannel: op}
				qc1._init()
				poisson1 := rng.NewPoissonGenerator(seed)
				ls1 := ls{name: "Alice.lightSource", timeline: &tl, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc1, poisson: poisson1}
				ls1._init()
				ls1.bs = MakeStateList()
				qsd1 := qsd{name: "bob.qsdetector", timeline: &tl2}
				qc1.receiver = &qsd1

				op2 := OpticalChannel{polarizationFidelity: 0.99, attenuation: 0.0002, distance: 10 * math.Pow10(3)}
				op2._init()
				qc2 := qc{name: "qc", timeline: &tl, OpticalChannel: op}
				qc2._init()
				poisson2 := rng.NewPoissonGenerator(seed)
				ls2 := ls{name: "Alice.lightSource", timeline: &tl2, frequency: 80 * math.Pow10(6), meanPhotonNum: 0.1, directReceiver: &qc2, poisson: poisson2}
				ls2._init()
				ls2.bs = MakeStateList()
				qsd2 := qsd{name: "bob.qsdetector", timeline: &tl}
				qc2.receiver = &qsd2

				message := kernel.Message{}
				process := kernel.Process{
					Fnptr:   ls1.transmit,
					Message: message,
					Owner:   ls1.timeline,
				}
				event := kernel.Event{
					Time:     0,
					Priority: 0,
					Process:  &process,
				}

				message2 := kernel.Message{}
				process2 := kernel.Process{
					Fnptr:   ls2.transmit,
					Message: message2,
					Owner:   ls2.timeline,
				}
				event2 := kernel.Event{
					Time:     0,
					Priority: 0,
					Process:  &process2,
				}

				tl.Schedule(&event)
				tl2.Schedule(&event2)

				kernel.Run([]*kernel.Timeline{&tl, &tl2})
				//fmt.Println(ls1.tran)
				t := time.Now()
				elapsed := t.Sub(start)
				fmt.Println(elapsed)
				tmp[1] = float64(elapsed)
			}
		}
		fmt.Println(fmt.Sprintf("%f",tmp[0]/tmp[1]))
	}
}
*/
