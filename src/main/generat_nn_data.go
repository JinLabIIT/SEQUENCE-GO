package main

import (
	"os"
	"quantum"
	"strconv"
)

func main() {
	path := os.Args[1]
	seed1, _ := strconv.Atoi(os.Args[2])
	seed2, _ := strconv.Atoi(os.Args[3])

	//fmt.Println(base_seed)
	quantum.RandGraphNN(path, seed1, seed2)
}
