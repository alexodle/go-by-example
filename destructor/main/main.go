package main

import (
	"flag"
	"fmt"
	"github.com/alexodle/go-by-example/destructor"
	"os"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Printf("Usage: main.go <input_dir> <output_dir>\n")
		os.Exit(1)
	}

	input_dir, output_dir := args[0], args[1]
	destructor.GenerateWrappers(input_dir, output_dir)
}
