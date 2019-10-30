package main

import (
	"fmt"
	"runtime/debug"
	"time"
)

func main() {
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Println(getID())
			debug.PrintStack()
		}()
	}
	time.Sleep(1 * time.Second)
}
