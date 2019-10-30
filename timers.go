package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTimer(1 * time.Second)

	go func() {
		fmt.Printf("a1\n")
		<-t.C
		fmt.Printf("a2\n")
	}()

	t2 := time.AfterFunc(1*time.Second, func() {
		fmt.Printf("b1\n")
	})

	t.Stop()
	t2.Stop()

	time.Sleep(3 * time.Second)
	fmt.Printf("DONE")
}
