package main

import "fmt"

func main() {
	m2 := map[interface{}]int{}

	m2[(&K1{}).Key()] = 1
	m2[(&K2{}).Key()] = 2
	m2[(&K3{}).Key()] = 3

	fmt.Println(m2)
}

type Keyer interface {
	Key() interface{}
}

type K1 struct{}

func (k *K1) Key() interface{} {
	return "hi"
}

type K2 struct{}

func (k *K2) Key() interface{} {
	return [2]string{"hi", "hello"}
}

type K3 struct{}

func (k *K3) Key() interface{} {
	return [2]string{"hi", "hello"}
}
