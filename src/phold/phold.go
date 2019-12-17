package main

import (
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
	"kernel"
	"math/rand"
	"strconv"
)

type node struct {
	name       string
	timeline   *kernel.Timeline
	totalNodes int
	otherNode  []*node
}

func (Node *node) nodeInit(timeline *kernel.Timeline, totalNodes int, name string) {
	Node.timeline = timeline
	Node.totalNodes = totalNodes
	Node.name = name
}

func initEvent(Node *node) {
	message := kernel.Message{"receiver": Node}
	process := kernel.Process{Fnptr: Node.send, Message: message, Owner: Node.timeline}
	delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
	event := kernel.Event{Time: delay, Process: &process, Priority: 0}
	Node.timeline.Schedule(&event)
}

func createEvent(Node *node, message kernel.Message, priority uint) kernel.Event {
	process := kernel.Process{Fnptr: Node.send, Message: message, Owner: Node.timeline}
	delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
	event := kernel.Event{Time: Node.timeline.LookAhead + Node.timeline.Now() + delay, Process: &process, Priority: priority}
	return event
}

func (Node *node) send(message kernel.Message) {
	receiver := message["receiver"].(node)
	target := rand.Intn(Node.totalNodes)
	newMessage := kernel.Message{"receiver": Node.otherNode[target]}
	event := createEvent(&receiver, newMessage, 0)
	Node.timeline.Schedule(&event)
}

func main() {
	// phold experience
	fmt.Println("phold experience")
	n := 1
	totalNodes := 4 // n <= totalentity
	endTime := uint64(1000)
	tl := make([]*kernel.Timeline, n)
	for i := 0; i < n; i++ {
		tl[i] = &kernel.Timeline{}
		tl[i].LookAhead = 100
		tl[i].Name = "Timeline " + strconv.Itoa(i) //covert int to string
		tl[i].SetEndTime(endTime)
	}

	nodeList := make([]*node, totalNodes)
	for i := 0; i < totalNodes; i++ {
		nodeList[i] = &node{}
		nodeList[i].otherNode = nodeList
		nodeList[i].nodeInit(tl[i%n], totalNodes, "Node"+strconv.Itoa(i))
	}
	phold(10000000, totalNodes, nodeList)
	kernel.Run(tl)
}

func phold(initJobs, totalNodes int, nodeList []*node) {
	for i := 0; i < initJobs; i++ {
		target := rand.Intn(totalNodes)
		initEvent(nodeList[target])
	}
}
