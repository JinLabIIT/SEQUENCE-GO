package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
)

type Thread struct {
	name string
}

func test(totalThreads, totalJobs int) {
	n := totalJobs / totalThreads
	var wg sync.WaitGroup
	for i := 0; i < totalThreads; i++ {
		wg.Add(1)
		thread := Thread{"thread" + strconv.Itoa(i)}
		go thread.addArray(n, &wg)
	}
	wg.Wait()
}

func (thread *Thread) addArray(n int, wg *sync.WaitGroup) {
	var i int
	list := make([]int, 10000)
	for i < n {
		for j := 0; j < 10000; j++ {
			list = append(list, i)
		}
		list = make([]int, 10000)
		i++
	}
	wg.Done()
}

func (thread *Thread) addOne(n int, wg *sync.WaitGroup) {
	var i int
	for i < n {
		i = i + 1
	}
	wg.Done()
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	totalThreads, _ := strconv.Atoi(os.Args[1])
	totalJobs, _ := strconv.Atoi(os.Args[2])
	past := time.Now()
	test(totalThreads, totalJobs)
	now := time.Now()
	fmt.Println("totalThreads is:", totalThreads, "Total consumption time is:", now.Sub(past))
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
