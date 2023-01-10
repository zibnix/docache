package docache

type Doer[T any] interface {
	Do() (T, error)
}

// The DoerFunc type is an adapter to allow the use of
// ordinary functions as Doers. If f is a function
// with the appropriate signature, DoerFunc[T](f) is a
// Doer that calls f.
type DoerFunc[T any] func() (T, error)

func (f DoerFunc[T]) Do() (T, error) {
	return f()
}
