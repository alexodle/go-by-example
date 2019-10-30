package a

import orig_a "github.com/alexodle/go-by-example/destructor/testdata/input/a"
import ab "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/a/ab"

type A interface {
	GetImpl() *orig_a.A
	F1(a2 A, b ab.AB, c orig_a.ANoMethods, d string) (*string, error)
}

func NewA(impl *orig_a.A) A {
	return &aWrapper{impl: impl}
}

type aWrapper struct {
	impl *orig_a.A
}

func (o *aWrapper) GetImpl() *orig_a.A {
	return o.impl
}

func (o *aWrapper) F1(a2 A, b ab.AB, c orig_a.ANoMethods, d string) (*string, error) {
	return o.impl.F1(*a2.GetImpl(), *b.GetImpl(), c, d)
}


