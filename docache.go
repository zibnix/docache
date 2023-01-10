package docache

import (
	"context"
	"log"
	"time"
)

// Call Loop on a goroutine, it does block.
// Loop returns immediately if the cache is already looping.
// Call Stop to stop looping.
// A Cache can be started and stopped and started again
// as long as the context provided to NewCache is not done.
// Data returns a slice, copied from the internal cache.
// The slice will be of length [0, capacity].
type Cache[T any] interface {
	Loop()
	Data() []Data[T]
	Stop()
}

// Data has a few extra fields added alongside your type.
// If your Doer returns an error, it is recorded in the
// Error field for that piece of Data.
type Data[T any] struct {
	Value     T
	Timestamp time.Time
	Error     error
}

// If you use a reference type, and modify the fields/contents
// after retrieving with the Data method, those changes will appear
// in the data when retrieved again from the Data method
// (assuming those data have not been pushed out by reaching max capacity).
// If that's a problem, don't change things after retrieval, or don't use
// a reference, or make a deeper copy of the data before mutating.
func NewCache[T any](
	ctx context.Context,
	interval time.Duration,
	capacity int,
	logger log.Logger,
	doer Doer[T]) Cache[T] {

	return &cache[T]{
		ctx:      ctx,
		interval: interval,
		capacity: capacity,
		logger:   logger,
		doer:     doer,
		data:     make([]Data[T], 0, capacity),
	}
}
