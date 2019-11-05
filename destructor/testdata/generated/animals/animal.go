package animals

import dog "github.com/alexodle/go-by-example/destructor/testdata/generated/animals/dog"
import orig_animals "github.com/alexodle/go-by-example/destructor/testdata/input/animals"
import orig_context "context"
import orig_dog "github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"

type AnimalDescription interface {
	GetImpl() *orig_animals.AnimalDescription
	GetBreed() string
	GetName() string
	GetWeight() int
	SetBreed(v string)
	SetName(v string)
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
	retval := o.impl.Breed
	return retval
}

func (o *animalDescriptionWrapper) SetBreed(v string) {
	o.impl.Breed = v
}

func (o *animalDescriptionWrapper) GetName() string {
	retval := o.impl.Name
	return retval
}

func (o *animalDescriptionWrapper) SetName(v string) {
	o.impl.Name = v
}

func (o *animalDescriptionWrapper) GetWeight() int {
	retval := o.impl.Weight
	return retval
}

func (o *animalDescriptionWrapper) SetWeight(v int) {
	o.impl.Weight = v
}

type Animals interface {
	GetImpl() *orig_animals.Animals
	AddAnimals(animals *[]interface{}) error
	AddDogs(dogs []dog.Dog) map[string]Animals
	GetAllDogs(ctx orig_context.Context) []dog.Dog
	GetAnimalDescription() AnimalDescription
	GetDogByName(name string) dog.Dog
	GetDogs() *[]dog.Dog
	GetDogsByNameField() *map[string]dog.Dog
	GetDogsByNames(names []string) map[string]dog.Dog
	GetLocations() []orig_animals.Location
	SetAnimalDescription(v AnimalDescription)
	SetDogs(v *[]dog.Dog)
	SetDogsByNameField(v *map[string]dog.Dog)
	SetLocations(v []orig_animals.Location)
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
	retval := o.impl.Locations
	return retval
}

func (o *animalsWrapper) SetLocations(v []orig_animals.Location) {
	o.impl.Locations = v
}

func (o *animalsWrapper) GetDogs() *[]dog.Dog {
	retval := o.impl.Dogs
	var retval_1 *[]dog.Dog
	if retval != nil {
		for _, it := range *retval {
			it_1 := dog.NewDog(&it)
			*retval_1 = append(*retval_1, it_1)
		}
	}
	return retval_1
}

func (o *animalsWrapper) SetDogs(v *[]dog.Dog) {
	var v_1 *[]orig_dog.Dog
	if v != nil {
		for _, it := range *v {
			it_1 := *it.GetImpl()
			*v_1 = append(*v_1, it_1)
		}
	}
	o.impl.Dogs = v_1
}

func (o *animalsWrapper) GetDogsByNameField() *map[string]dog.Dog {
	retval := o.impl.DogsByNameField
	var retval_1 *map[string]dog.Dog
	if retval != nil {
		retval_1 = &map[string]dog.Dog{}
		for k, it := range *retval {
			it_1 := dog.NewDog(&it)
			(*retval_1)[k] = it_1
		}
	}
	return retval_1
}

func (o *animalsWrapper) SetDogsByNameField(v *map[string]dog.Dog) {
	var v_1 *map[string]orig_dog.Dog
	if v != nil {
		v_1 = &map[string]orig_dog.Dog{}
		for k, it := range *v {
			it_1 := *it.GetImpl()
			(*v_1)[k] = it_1
		}
	}
	o.impl.DogsByNameField = v_1
}

func (o *animalsWrapper) GetAnimalDescription() AnimalDescription {
	retval := o.impl.AnimalDescription
	retval_1 := NewAnimalDescription(retval)
	return retval_1
}

func (o *animalsWrapper) SetAnimalDescription(v AnimalDescription) {
	v_1 := v.GetImpl()
	o.impl.AnimalDescription = v_1
}

func (o *animalsWrapper) GetAllDogs(ctx orig_context.Context) []dog.Dog {
	retval0 := o.impl.GetAllDogs(ctx)
	var retval0_1 []dog.Dog
	for _, it := range retval0 {
		it_1 := dog.NewDog(it)
		retval0_1 = append(retval0_1, it_1)
	}
	return retval0_1
}

func (o *animalsWrapper) GetDogsByNames(names []string) map[string]dog.Dog {
	retval0 := o.impl.GetDogsByNames(names)
	var retval0_1 map[string]dog.Dog
	retval0_1 = map[string]dog.Dog{}
	for k, it := range retval0 {
		it_1 := dog.NewDog(it)
		retval0_1[k] = it_1
	}
	return retval0_1
}

func (o *animalsWrapper) GetDogByName(name string) dog.Dog {
	retval0 := o.impl.GetDogByName(name)
	retval0_1 := dog.NewDog(&retval0)
	return retval0_1
}

func (o *animalsWrapper) AddAnimals(animals *[]interface{}) error {
	retval0 := o.impl.AddAnimals(animals)
	return retval0
}

func (o *animalsWrapper) AddDogs(dogs []dog.Dog) map[string]Animals {
	var dogs_1 []orig_dog.Dog
	for _, it := range dogs {
		it_1 := *it.GetImpl()
		dogs_1 = append(dogs_1, it_1)
	}
	retval0 := o.impl.AddDogs(dogs_1)
	var retval0_1 map[string]Animals
	retval0_1 = map[string]Animals{}
	for k, it := range retval0 {
		it_1 := NewAnimals(&it)
		retval0_1[k] = it_1
	}
	return retval0_1
}
