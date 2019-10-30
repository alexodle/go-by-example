package main

import "fmt"

type person struct {
	name string
	age  int
}

func main() {
	p1 := person{name: "allen", age: 20}
	p1_cp := p1
	p1_ptr := &p1
	fmt.Println("p1:", p1)
	fmt.Println("p1_cp:", p1_cp)
	fmt.Println("p1_ptr:", p1_ptr)
	fmt.Println("")

	p1_cp.age = 21
	fmt.Println("p1:", p1)
	fmt.Println("p1_cp:", p1_cp)
	fmt.Println("p1_ptr:", p1_ptr)
	fmt.Println("")

	p1_ptr.age = 22
	fmt.Println("p1:", p1)
	fmt.Println("p1_cp:", p1_cp)
	fmt.Println("p1_ptr:", p1_ptr)
	fmt.Println("")
}
