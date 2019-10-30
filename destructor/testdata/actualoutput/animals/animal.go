package animals

import dog "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals/dog"
import orig_animals "github.com/alexodle/go-by-example/destructor/testdata/input/animals"

type Animals interface {
	GetImpl() *orig_animals.Animals
	GetLocations() []orig_animals.Location
	SetLocations(v []orig_animals.Location)
	GetAllDogs() []*dog.Dog
}

func NewAnimals(impl *orig_animals.Animals) Animals {
	return &animalsWrapper{impl: impl}
}

type animalsWrapper struct {
	impl *orig_animals.Animals
}

func (o *animalsWrapper) GetImpl() *orig_animals.Animals {
	return o.impl
}

func (o *animalsWrapper) GetLocations() []orig_animals.Location {
	return o.impl.Locations
}

func (o *animalsWrapper) SetLocations(v []orig_animals.Location) {
	o.impl.Locations = v
}

func (o *animalsWrapper) GetAllDogs() []*dog.Dog {
	return o.impl.GetAllDogs()
}
