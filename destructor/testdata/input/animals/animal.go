package animals

import (
	"context"
	"github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"
)

type Animals struct {
	Locations []Location
	*AnimalDescription
}

func (a *Animals) GetAllDogs(ctx context.Context) []*dog.Dog {
	return nil
}

func (a *Animals) GetDogsByNames(names []string) map[string]*dog.Dog {
	return nil
}

func (a *Animals) GetDogByName(name string) dog.Dog {
	return dog.Dog{}
}

func (a *Animals) AddAnimals(animals *[]interface{}) error {
	return nil
}

type AnimalDescription struct {
	Breed string
	Name string
	Weight int
}

type Location struct {}
