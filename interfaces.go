package main

import (
	"fmt"
	"math"
)

type geometry interface {
	area() float64
	perim() float64
}

type rect struct {
	width, height float64
}

type circle struct {
	radius float64
}

func (r rect) area() float64 {
	return r.height * r.width
}
func (r rect) perim() float64 {
	return r.height*2 + r.width*2
}

func (c circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

func measure(g geometry) {
	fmt.Println("g:", g)
	fmt.Println("g.area():", g.area())
	fmt.Println("g.perim():", g.perim())
}

func main() {
	measure(rect{width: 5, height: 10})
	// measure(circle{radius: 5})
}
