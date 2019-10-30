package a

import (
	"github.com/alexodle/go-by-example/destructor/testdata/input/a/ab"
)

type ANoMethods struct {
	V1 string
}

type A struct {
	V1 string
}

func (a *A) f1(b, c string) (string, error) {
	return "", nil
}

func (a *A) F1(a2 A, b *ab.AB, c ANoMethods, d string) (*string, error) {
	s := ""
	return &s, nil
}

