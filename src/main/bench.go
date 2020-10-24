package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"quantum"
	"time"
)

func main() {
	f, err := os.OpenFile("main.log", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for i := 1; i <= 32; i *= 2 {
		tick := time.Now()
		quantum.Main(128, i, uint64(5*math.Pow10(7)), uint64(math.Pow10(11)))
		tock := time.Since(tick)
		s := fmt.Sprintf("%d %d\n", i, int(tock))
		d1 := []byte(s)
		f.Write(d1)
	}

}
