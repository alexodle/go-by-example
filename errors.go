package main

import (
	"errors"
	"fmt"
)

func f1(arg int) (int, error) {
	if arg == 42 {
		return -1, errors.New("I don't do 42s")
	}
	return arg + 3, nil
}

type argError struct {
	arg  int
	prob string
}

func (e *argError) Error() string {
	return fmt.Sprintf("%d - %s", e.arg, e.prob)
}

func f2(arg int) (int, error) {
	if arg == 42 {
		return -1, &argError{arg, "can't with it"}
	}
	return arg + 4, nil
}

func main() {
	if r, e := f1(42); e != nil {
		fmt.Println("e:", e)
	} else {
		fmt.Println("r:", r)
	}

	if r, e := f2(42); e != nil {
		fmt.Println("e:", e)
	} else {
		fmt.Println("r:", r)
	}
}
