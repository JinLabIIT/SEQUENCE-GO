package main

import (
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
	"kernel"
	"math/rand"
	"strconv"
)

type Node struct {
	name       string
	timeline   *kernel.Timeline
	totalNodes int
	otherNode  []*Node
}

func (node *Node) nodeInit(timeline *kernel.Timeline, totalNodes int, name string) {
	node.timeline = timeline
	node.totalNodes = totalNodes
	node.name = name
}

func setSeed(seed int64) {
	rand.Seed(seed)
}

func initEvent(node *Node) {
	message := kernel.Message{"receiver": node}
	process := kernel.Process{Fnptr: node.send, Message: message, Owner: node.timeline}
	delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
	event := kernel.Event{Time: delay, Process: &process, Priority: 0}
	node.timeline.Schedule(&event)
}

func createEvent(node *Node, message kernel.Message, priority uint) kernel.Event {
	process := kernel.Process{Fnptr: node.send, Message: message, Owner: node.timeline}
	delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
	event := kernel.Event{Time: node.timeline.LookAhead + node.timeline.Now() + delay, Process: &process, Priority: priority}
	return event
}

func (node *Node) send(message kernel.Message) {
	receiver := message["receiver"].(*Node)
	target := rand.Intn(node.totalNodes)
	newMessage := kernel.Message{"receiver": node.otherNode[target]}
	event := createEvent(receiver, newMessage, 0)
	node.timeline.Schedule(&event)
}

func main() {
	// phold experience
	fmt.Println("phold experience")
	seed := int64(12345)
	setSeed(seed)
	totalThreads := 1
	totalNodes := 4 // totalThreads <= totalNodes
	endTime := uint64(1000)
	initJobs := 1000000
	lookAhead := uint64(100)
	phold(initJobs, totalThreads, totalNodes, endTime, lookAhead)
}

func phold(initJobs, totalThreads, totalNodes int, endTime, lookAhead uint64) {
	tl := make([]*kernel.Timeline, totalThreads)
	for i := 0; i < totalThreads; i++ {
		tl[i] = &kernel.Timeline{}
		tl[i].Init(lookAhead, endTime)
		tl[i].Name = "Timeline " + strconv.Itoa(i) //covert int to string
		tl[i].SetEndTime(endTime)
	}

	nodeList := make([]*Node, totalNodes)
	for i := 0; i < totalNodes; i++ {
		nodeList[i] = &Node{}
		nodeList[i].otherNode = nodeList
		nodeList[i].nodeInit(tl[i%totalThreads], totalNodes, "Node"+strconv.Itoa(i))
	}
	for i := 0; i < initJobs; i++ {
		target := rand.Intn(totalNodes)
		initEvent(nodeList[target])
	}
	kernel.Run(tl)
}
