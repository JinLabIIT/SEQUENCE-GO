package main

import (
	"os"
	"quantum"
	"strconv"
)

func main() {
	threadNum, _ := strconv.Atoi(os.Args[1])
	repeat, _ := strconv.Atoi(os.Args[2])
	filename := os.Args[3]
	logPath := os.Args[4]
	optimization := os.Args[5]

	//fmt.Println(base_seed)
	if optimization == "0" {
		quantum.RandGraph(threadNum, repeat, filename, logPath, false)
	} else {
		quantum.RandGraph(threadNum, repeat, filename, logPath, true)
	}

}
