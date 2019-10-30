package a

import orig_a "github.com/alexodle/go-by-example/destructor/testdata/input/a"
import ab "github.com/alexodle/go-by-example/destructor/testdata/actualoutput/a/ab"

type A interface {
	GetImpl() *orig_a.A
	GetV1() (string)
	SetV1(v string)
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

func (o *aWrapper) GetV1() (string) {
	return o.impl.V1
}

func (o *aWrapper) SetV1(v string) {
	o.impl.V1 = v
}

func (o *aWrapper) F1(a2 A, b ab.AB, c orig_a.ANoMethods, d string) (*string, error) {
	return o.impl.F1(*a2.GetImpl(), *b.GetImpl(), c, d)
}


