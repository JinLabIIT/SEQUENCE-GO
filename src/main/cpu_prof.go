package main

import (
	"math"
	"os"
	"quantum"
	"runtime/pprof"
	"strconv"
)

func main() {
	filename := os.Args[1]
	file, _ := os.Create(filename)
	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile()

	_threadNum := os.Args[2]
	threadNum, _ := strconv.Atoi(_threadNum)

	quantum.Main(128, threadNum, uint64(5*math.Pow10(7)), uint64(math.Pow10(10)))
}
