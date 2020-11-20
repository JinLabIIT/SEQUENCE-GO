package main

import (
	"os"
	"quantum"
	"strconv"
)

func main() {
	path := os.Args[1]
	base_seed, _ := strconv.Atoi(os.Args[2])

	//fmt.Println(base_seed)
	quantum.RandGraphNN(path, base_seed)
}
