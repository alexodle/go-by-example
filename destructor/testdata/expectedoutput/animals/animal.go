package animals

import dog "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/animals/dog"
import orig_animals "github.com/alexodle/go-by-example/destructor/testdata/input/animals"
import orig_context "context"
import orig_dog "github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"

type Animals interface {
	GetImpl() *orig_animals.Animals
	GetLocations() []orig_animals.Location
	SetLocations(v []orig_animals.Location)
	GetAnimalDescription() AnimalDescription
	SetAnimalDescription(v AnimalDescription)
	GetAllDogs(ctx orig_context.Context) []dog.Dog
	GetDogsByNames(names []string) map[string]dog.Dog
	GetDogByName(name string) dog.Dog
	AddAnimals(animals *[]interface{}) error
	AddDogs(dogs []dog.Dog) map[string]Animals
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

func (o *animalsWrapper) GetAnimalDescription() AnimalDescription {
	v0 := o.impl.AnimalDescription
	newv0 := NewAnimalDescription(v0)
	return newv0
}

func (o *animalsWrapper) SetAnimalDescription(v AnimalDescription) {
	newv := v.GetImpl()
	o.impl.AnimalDescription = newv
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

func (o *animalsWrapper) AddAnimals(animals *[]interface{}) error {
	v0 := o.impl.AddAnimals(animals)
	return v0
}

func (o *animalsWrapper) AddDogs(dogs []dog.Dog) map[string]Animals {
	var newdogs []orig_dog.Dog
	for _, v := range dogs {
		newdogs = append(newdogs, *v.GetImpl())
	}
	v0 := o.impl.AddDogs(newdogs)
	var newv0 map[string]Animals
	for k, v := range v0 {
		newv0[k] = NewAnimals(&v)
	}
	return newv0
}

type AnimalDescription interface {
	GetImpl() *orig_animals.AnimalDescription
	GetBreed() string
	SetBreed(v string)
	GetName() string
	SetName(v string)
	GetWeight() int
	SetWeight(v int)
}

func NewAnimalDescription(impl *orig_animals.AnimalDescription) AnimalDescription {
	return &animalDescriptionWrapper{impl: impl}
}

type animalDescriptionWrapper struct {
	impl *orig_animals.AnimalDescription
}

func (o *animalDescriptionWrapper) GetImpl() *orig_animals.AnimalDescription {
	return o.impl
}

func (o *animalDescriptionWrapper) GetBreed() string {
	v0 := o.impl.Breed
	return v0
}

func (o *animalDescriptionWrapper) SetBreed(v string) {
	o.impl.Breed = v
}

func (o *animalDescriptionWrapper) GetName() string {
	v0 := o.impl.Name
	return v0
}

func (o *animalDescriptionWrapper) SetName(v string) {
	o.impl.Name = v
}

func (o *animalDescriptionWrapper) GetWeight() int {
	v0 := o.impl.Weight
	return v0
}

func (o *animalDescriptionWrapper) SetWeight(v int) {
	o.impl.Weight = v
}
