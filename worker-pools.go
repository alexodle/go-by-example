package main

import (
	"fmt"
	"math/rand"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Printf("Worker %d received job %d\n", id, j)
		r := rand.Intn(1000)
		time.Sleep(time.Duration(r) * time.Millisecond)
		results <- -j
		fmt.Printf("Worker %d processed job %d\n", id, j)
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int)

	for i := 0; i < 8; i++ {
		go worker(i, jobs, results)
	}

	for i := 0; i < 100; i++ {
		jobs <- i
	}
	fmt.Println("Finished queueing jobs")
	close(jobs)

	/*fmt.Println("Awaiting results")
	for i := 0; i < 100; i++ {
		r := <- results
		fmt.Println("Result received", r)
	}*/
	fmt.Println("Done")
}
