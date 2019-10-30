package main

import (
	"fmt"
	"math/rand"
	"time"
)

type readOp struct {
	key  int
	resp chan int
}
type writeOp struct {
	key  int
	val  int
	resp chan bool
}
type stats struct {
	reads  uint64
	writes uint64
}
type statsOp struct {
	resp chan stats
}

func main() {

	writes := make(chan *writeOp)
	reads := make(chan *readOp)
	readStats := make(chan *statsOp)

	go func() {
		state := make(map[int]int)
		stats := stats{}
		for {
			select {
			case read := <-reads:
				//fmt.Printf("Reading key: %d -> %d\n", read.key, state[read.key])
				read.resp <- state[read.key]
				stats.reads++
			case write := <-writes:
				//fmt.Printf("Writing key: %d -> %d\n", write.key, write.val)
				state[write.key] = write.val
				write.resp <- true
				stats.writes++
			case readStat := <-readStats:
				readStat.resp <- stats
			}
		}
	}()

	for i := 0; i < 100; i++ {
		go func() {
			for {
				read := &readOp{key: rand.Intn(100), resp: make(chan int)}
				reads <- read
				<-read.resp
				time.Sleep(time.Millisecond)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for {
				write := &writeOp{key: rand.Intn(100), val: rand.Intn(1000), resp: make(chan bool)}
				writes <- write
				<-write.resp
				time.Sleep(time.Millisecond)
			}
		}()
	}

	time.Sleep(1 * time.Second)
	statsReq := &statsOp{resp: make(chan stats)}
	readStats <- statsReq
	stats := <-statsReq.resp
	fmt.Println("final stats:", stats)
}
