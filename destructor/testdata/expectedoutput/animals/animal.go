package animals

import dog "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals/dog"
import orig_animals "github.com/alexodle/go-by-example/destructor/testdata/input/animals"
import orig_context "context"

type Animals interface {
	GetImpl() *orig_animals.Animals
	GetLocations() []orig_animals.Location
	SetLocations(v []orig_animals.Location)
	GetAnimalDescription() *orig_animals.AnimalDescription
	SetAnimalDescription(v *orig_animals.AnimalDescription)
	GetAllDogs(ctx orig_context.Context) []dog.Dog
	GetDogsByNames(names []string) map[string]dog.Dog
	GetDogByName(name string) dog.Dog
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
	v0 := o.impl.Locations
	return v0
}

func (o *animalsWrapper) SetLocations(v []orig_animals.Location) {
	o.impl.Locations = v
}

func (o *animalsWrapper) GetAnimalDescription() *orig_animals.AnimalDescription {
	v0 := o.impl.AnimalDescription
	return v0
}

func (o *animalsWrapper) SetAnimalDescription(v *orig_animals.AnimalDescription) {
	o.impl.AnimalDescription = v
}

func (o *animalsWrapper) GetAllDogs(ctx orig_context.Context) []dog.Dog {
	v0 := o.impl.GetAllDogs(ctx)
	var newv0 []dog.Dog
	for _, v := range v0 {
		newv0 = append(newv0, dog.NewDog(v))
	}
	return newv0
}

func (o *animalsWrapper) GetDogsByNames(names []string) map[string]dog.Dog {
	v0 := o.impl.GetDogsByNames(names)
	var newv0 map[string]dog.Dog
	for k, v := range v0 {
		newv0[k] = dog.NewDog(v)
	}
	return newv0
}

func (o *animalsWrapper) GetDogByName(name string) dog.Dog {
	v0 := o.impl.GetDogByName(name)
	newv0 := dog.NewDog(&v0)
	return newv0
}
