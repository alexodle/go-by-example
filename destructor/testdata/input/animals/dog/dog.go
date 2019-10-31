package dog

import (
	"github.com/alexodle/go-by-example/destructor/testdata/input/animals"
	"github.com/alexodle/go-by-example/destructor/testdata/input/animals/food"
)

type Dog struct {
}

func (d *Dog) Describe() animals.AnimalDescription {
	return animals.AnimalDescription{}
}

// Callback functions not supported yet, should just leave out
func (d *Dog) OnBark(s1 string, barkHandler func(d *Dog, s string) string) {
}

func (d *Dog) Barks() bool {
	return true
}

func (d *Dog) Meows() bool {
	return false
}

func (d *Dog) Eat(f *food.Food) int {
	return 1
}

func (d *Dog) Clone() (*Dog, error) {
	return nil, nil
}
