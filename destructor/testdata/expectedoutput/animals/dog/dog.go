package dog

import food "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals/food"
import orig_animals "github.com/alexodle/go-by-example/destructor/testdata/input/animals"
import orig_dog "github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"

type Dog interface {
	GetImpl() *orig_dog.Dog
	Describe() orig_animals.AnimalDescription
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

func (o *dogWrapper) Describe() orig_animals.AnimalDescription {
	return o.impl.Describe()
}

func (o *dogWrapper) Barks() bool {
	return o.impl.Barks()
}

func (o *dogWrapper) Meows() bool {
	return o.impl.Meows()
}

func (o *dogWrapper) Eat(f food.Food) int {
	return o.impl.Eat(f.GetImpl())
}

func (o *dogWrapper) Clone() (Dog, error) {
	v, err := o.impl.Clone()
	if err != nil {
		return nil, err
	}
	return NewDog(v), nil
}
