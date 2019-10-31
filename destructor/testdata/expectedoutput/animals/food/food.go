package food

import orig_food "github.com/alexodle/go-by-example/destructor/testdata/input/animals/food"

type Food interface {
	GetImpl() *orig_food.Food
	Brand() string
}

func NewFood(impl *orig_food.Food) Food {
	return &foodWrapper{impl: impl}
}

type foodWrapper struct {
	impl *orig_food.Food
}

func (o *foodWrapper) GetImpl() *orig_food.Food {
	return o.impl
}

func (o *foodWrapper) Brand() string {
	v0 := o.impl.Brand()
	return v0
}
