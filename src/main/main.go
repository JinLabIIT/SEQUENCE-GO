package main

import (
	"fmt"
	"github.com/pkg/profile"
	"os"
	"quantum"
	"strconv"
	"time"
)

func main() {
	//a()
	//b()
	c()
}
func a() {
	defer profile.Start().Stop()
	totalNodes, _ := strconv.Atoi(os.Args[1])
	totalThreads, _ := strconv.Atoi(os.Args[2])
	past := time.Now()
	quantum.Main(totalNodes, totalThreads, 50000000)
	now := time.Now()
	fmt.Println("totalThreads is:", totalThreads, "Total consumption time is:", now.Sub(past))
}

func b() {
	//defer profile.Start().Stop()
	totalThreads, _ := strconv.Atoi(os.Args[1])
	quantum.RandGraph(totalThreads, "../../tools/1.json", false)
}

func c() {
	//defer profile.Start().Stop()
	//fmt.Println("hello world")
	quantum.BB84Test()
}
