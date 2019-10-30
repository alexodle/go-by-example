package reftest

import "github.com/alexodle/go-by-example/reftest/innerreftest"

type AA struct {}

type A struct {
}

//func (a *A) f1(b, c string) (string, error) {
//	return "", nil
//}
//
//func (a *A) F2(b, c string) (string, error) {
//	return "", nil
//}
//
//func (a A) F3(b, c string, d *string) (string, error) {
//	return "", nil
//}

func (a A) F4(a2 A, b innerreftest.B, c AA) (string, error) {
	return "", nil
}
