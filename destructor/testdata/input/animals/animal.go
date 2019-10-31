package animals

import (
	"context"
	"github.com/alexodle/go-by-example/destructor/testdata/input/animals/dog"
)

type Animals struct {
	Locations []Location
}

func (a *Animals) GetAllDogs(ctx context.Context) []*dog.Dog {
	return nil
}

func (a *Animals) GetDogsByName() map[string]*dog.Dog {
	return nil
}

type AnimalDescription struct {
	Breed string
	Name string
	Weight int
}

type Location struct {}
