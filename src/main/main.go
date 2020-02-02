package main

import (
	"fmt"
	"os"
	"quantum"
	"strconv"
	"time"
)

func main() {
	totalNodes, _ := strconv.Atoi(os.Args[1])
	totalThreads, _ := strconv.Atoi(os.Args[2])
	past := time.Now()
	quantum.Main(totalNodes, totalThreads, 50000000)
	now := time.Now()
	fmt.Println("totalThreads is:", totalThreads, "Total consumption time is:", now.Sub(past))
}
