package ab

import orig_ab "github.com/alexodle/go-by-example/destructor/testdata/input/a/ab"

type AB interface {
	GetImpl() *orig_ab.AB
	F1(b string, c string) (string, error)
}

func NewAB(impl *orig_ab.AB) AB {
	return &aBWrapper{impl: impl}
}

type aBWrapper struct {
	impl *orig_ab.AB
}

func (o *aBWrapper) GetImpl() *orig_ab.AB {
	return o.impl
}

func (o *aBWrapper) F1(b string, c string) (string, error) {
	return o.impl.F1(b, c)
}


