package main

import (
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
	"kernel"
	"math/rand"
	"strconv"
)

type node struct {
	name string
}

//func block_low(id,p,n int) int {return int(id*n/p)}
//func block_high(id,p,n int) int {return block_low(id+1,p,n)-1}
//func block_size(id,p,n int) int {return block_low(id+1,p,n)-block_low(id,p,n)}
//func block_owner(index, p, n int) int {return int((p*(index+1)-1)/n)}

func initEvent(sender *kernel.Entity, delay uint64, totalNodes int, timeline *kernel.Timeline, entityList []*kernel.Entity) {
	name := node{sender.Name}
	message := kernel.Message{"sender": sender, "totalNodes": totalNodes, "entityList": entityList}
	process := kernel.Process{Fnptr: name.send, Message: message, Owner: sender}
	event := kernel.Event{Time: delay, Process: &process}
	timeline.Schedule(&event)
}

func createEvent(sender, receiver *kernel.Entity, delay uint64, totalNodes int, timeline *kernel.Timeline, entityList []*kernel.Entity) {
	name := node{sender.Name}
	message := kernel.Message{"sender": sender, "totalNodes": totalNodes, "entityList": entityList}
	process := kernel.Process{Fnptr: name.send, Message: message, Owner: receiver}
	event := kernel.Event{Time: timeline.LookAhead + timeline.Now() + delay, Process: &process, Priority: 0}
	timeline.Schedule(&event)
}

func (node *node) send(message kernel.Message) {
	sender := message["sender"].(*kernel.Entity)
	totalNodes := message["totalNodes"].(int)
	entityList := message["entityList"].([]*kernel.Entity)
	delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
	target := rand.Intn(totalNodes)
	createEvent(sender, entityList[target], delay, totalNodes, sender.Timeline, entityList)
}

func main() {
	// phold experience
	fmt.Println("phold experience")
	n := 4
	totalEntity := 4 // n <= totalentity
	endTime := uint64(1000)
	tl := make([]*kernel.Timeline, n)
	for i := 0; i < n; i++ {
		tl[i] = &kernel.Timeline{}
		tl[i].LookAhead = 100
		tl[i].Name = "Timeline " + strconv.Itoa(i) //covert int to string
		tl[i].SetEndTime(endTime)
	}
	//tmp := 0
	entityList := make([]*kernel.Entity, totalEntity)
	for i := 0; i < totalEntity; i++ {
		entity := kernel.Entity{Timeline: tl[i%n]}
		entityList[i] = &entity
		tl[i%n].SetEntities(entity)
	}
	phold(10000000, n, entityList)
	kernel.Run(tl, n)
}

func phold(initJobs, totalNodes int, entityList []*kernel.Entity) {
	for i := 0; i < initJobs; i++ {
		target := rand.Intn(totalNodes)
		delay := uint64(distuv.Poisson{Lambda: 1}.Rand())
		initEvent(entityList[target], delay, totalNodes, entityList[target].Timeline, entityList)
	}
}
