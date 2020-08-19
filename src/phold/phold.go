package main

import (
	"fmt"
	rng "github.com/leesper/go_rng"
	"kernel"
	"os"
	_ "runtime/trace"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	name       string
	timeline   *kernel.Timeline
	totalNodes int
	otherNode  []*Node
	exp        *rng.ExpGenerator
	ung        *rng.UniformGenerator
}

func (node *Node) nodeInit(timeline *kernel.Timeline, totalNodes int, name string, seed int64) {
	node.timeline = timeline
	node.totalNodes = totalNodes
	node.name = name
	node.exp = rng.NewExpGenerator(seed)
	node.ung = rng.NewUniformGenerator(seed)
}

func initEvent(node *Node) {
	message := kernel.Message{"receiver": node}
	process := kernel.Process{
		Fnptr:   node.send,
		Message: message,
		Owner:   node.timeline,
	}
	event := kernel.Event{
		Time:     uint64(node.exp.Exp(1) * 100),
		Priority: 0,
		Process:  &process,
	}
	node.timeline.Schedule(&event)
}

func (node *Node) send(message kernel.Message) {
	receiver := message["receiver"].(*Node)
	target := node.ung.Int32Range(0, int32(node.totalNodes))
	t := node.timeline.LookAhead + node.timeline.Now() + uint64(node.exp.Exp(1)*100)
	event := node.timeline.EventPool.Get().(*kernel.Event)
	event.Time = t
	event.Priority = 0
	event.Process.Message["receiver"] = node.otherNode[target]
	event.Process.Fnptr = receiver.send
	event.Process.Owner = receiver.timeline
	node.timeline.Schedule(event)
}

func main() {
	//trace.Start(os.Stderr)
	//defer trace.Stop()
	//defer debug.SetGCPercent(debug.SetGCPercent(-1))
	//defer profile.Start().Stop()
	fmt.Println("phold simulation")
	seed := int64(123456)
	totalThreads, _ := strconv.Atoi(os.Args[1])
	totalNodes, _ := strconv.Atoi(os.Args[2]) // totalThreads <= totalNodes
	endTime := uint64(50000)
	initJobs := 100000
	lookAhead := uint64(100)
	phold(initJobs, totalThreads, totalNodes, endTime, lookAhead, seed)
}

func phold(initJobs, totalThreads, totalNodes int, endTime, lookAhead uint64, seed int64) {

	tl := make([]*kernel.Timeline, totalThreads)
	ung := rng.NewUniformGenerator(seed)
	count := 0
	for i := 0; i < totalThreads; i++ {
		tl[i] = &kernel.Timeline{}
		tl[i].Init(lookAhead, endTime)
		tl[i].Name = "Timeline" + strconv.Itoa(i) //covert int to string
		tl[i].SetEndTime(endTime)

		var eventPool = sync.Pool{
			New: func() interface{} {
				message := &kernel.Message{}
				process := &kernel.Process{}
				process.Message = *message
				event := &kernel.Event{}
				event.Process = process
				count++
				return event
			},
		}

		tl[i].EventPool = &eventPool
	}

	nodeList := make([]*Node, totalNodes)
	for i := 0; i < totalNodes; i++ {
		nodeList[i] = &Node{}
		nodeList[i].otherNode = nodeList
		nodeList[i].nodeInit(tl[i%totalThreads], totalNodes, "Node"+strconv.Itoa(i), seed)
	}
	for i := 0; i < initJobs; i++ {
		target := ung.Int32Range(0, int32(totalNodes))
		initEvent(nodeList[target])
	}
	past := time.Now()
	kernel.Run(tl)
	now := time.Now()
	fmt.Println("totalThreads is:", totalThreads, "Total consumption time is:", now.Sub(past))
	fmt.Println(count)
}
