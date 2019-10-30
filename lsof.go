package main

import (
	"fmt"
	"os/exec"
)

func lsof() {
	cmd := exec.Command("lsof", "hello.txt")
	cmd.Dir = "/Volumes/git/go/src/github.com/alexodle/go-by-example"

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("Error!", err.Error())
		fmt.Println(string(output))
	} else {
		fmt.Println(string(output))
	}
}

func main() {
	lsof()
}
