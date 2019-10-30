package main

import "fmt"

func main() {
	s := "hello there friend\n"
	fmt.Println(len(s), len([]byte(s)))
}
