package docache

import (
	"errors"
	"testing"
)

type MyDoer[T any] struct{}

func (m *MyDoer[T]) Do() (T, error) {
	return *new(T), nil
}

type IntDoer struct{}

func (i *IntDoer) Do() (int, error) {
	return 0, nil
}

func doer[T any]() (T, error) {
	return *new(T), nil
}

func intDoer() (int, error) {
	return 0, nil
}

// this test should fail to compile, rather than fail at runtime if something is wrong
func TestDoerTypeConversions(t *testing.T) {
	// make sure a *MyDoer[T] is assignable to a Doer[T]
	var ds Doer[string]
	m := &MyDoer[string]{}
	ds = m
	_ = ds

	// make sure an *IntDoer is assignable to a Doer[int]
	var di Doer[int]
	i := &IntDoer{}
	di = i

	// a variable cannot receive type parameters
	anon := func() (int, error) {
		return 0, errors.New("")
	}

	// but this anonymous/literal func should be convertible to a Doer[int]
	di = DoerFunc[int](anon)

	// and these named functions should be convertible to a Doer[int]
	di = DoerFunc[int](doer[int])
	di = DoerFunc[int](intDoer)

	_ = di
}
