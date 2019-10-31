package dog

import animals "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals"
import food "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals/food"

import orig_dog "github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"

type Dog interface {
	GetImpl() *orig_dog.Dog
	Describe() animals.AnimalDescription
	Barks() bool
	Meows() bool
	Eat(f food.Food) int
	Clone() (Dog, error)
}

func NewDog(impl *orig_dog.Dog) Dog {
	return &dogWrapper{impl: impl}
}

type dogWrapper struct {
	impl *orig_dog.Dog
}

func (o *dogWrapper) GetImpl() *orig_dog.Dog {
	return o.impl
}

func (o *dogWrapper) Describe() animals.AnimalDescription {
	v0 := o.impl.Describe()
	newv0 := animals.NewAnimalDescription(&v0)
	return newv0
}

func (o *dogWrapper) Barks() bool {
	v0 := o.impl.Barks()
	return v0
}

func (o *dogWrapper) Meows() bool {
	v0 := o.impl.Meows()
	return v0
}

func (o *dogWrapper) Eat(f food.Food) int {
	newf := f.GetImpl()
	v0 := o.impl.Eat(newf)
	return v0
}

func (o *dogWrapper) Clone() (Dog, error) {
	v0, v1 := o.impl.Clone()
	newv0 := NewDog(v0)
	return newv0, v1
}
