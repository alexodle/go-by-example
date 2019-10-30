package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := os.Stat("~/Desktop/scratch")
	if os.IsNotExist(err) {
		fmt.Printf("does opt!")
	}
}
