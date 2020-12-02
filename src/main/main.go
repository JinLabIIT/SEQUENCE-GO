package main

import (
	"fmt"
	"os"
	"quantum"
	"strconv"
	"time"
)

func main() {
	a()
	// b()
	//c()
}
func a() {
	totalNodes, _ := strconv.Atoi(os.Args[1])
	totalThreads, _ := strconv.Atoi(os.Args[2])
	past := time.Now()
	quantum.Main(totalNodes, totalThreads, 50000000, 16847704064)
	now := time.Now()
	fmt.Println("totalThreads is:", totalThreads, "Total consumption time is:", now.Sub(past))
}

func b() {
	totalThreads, _ := strconv.Atoi(os.Args[1])
	quantum.RandGraph(totalThreads, "../../tools/1.json", false)
}

func c() {
	//defer debug.SetGCPercent(debug.SetGCPercent(-1))
	quantum.BB84Test()
}
