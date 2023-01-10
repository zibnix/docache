# DoCache
Do something in a loop, caching the results. All methods safe for concurrent use.

### Why?

This code can allow you to control the frequency that some data is generated, rather
than responding to something that might have a non-ideal frequency, like network requests.
Use this code to prevent network activity from reaching your data storage or devices if
you don't ever want a request to wait for backend IO.

## The Cache[T] interface this pkg provides
```go
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
```

```go
// Data has a few extra fields added alongside your type.
// If your Doer returns an error, it is recorded in the
// Error field for that piece of Data.
type Data[T any] struct {
	Value     T
	Timestamp time.Time
	Error     error
}
```

## Implement the Doer[T] interface in your code
A `Doer` just returns any type T. This can be a method on a type you define,
or if you have a simple `Doer`, you can cast a `func` with a matching signature
to a `DoerFunc`. This is the same way you would use `http.HandlerFunc` from `net/http`.
```go
type Doer[T any] interface {
	Do() (T, error)
}

type DoerFunc[T any] func() (T, error)

```

## Use the constructor
```go
func NewCache[T any](ctx context.Context, interval time.Duration, capacity int, logger log.Logger, doer Doer[T]) Cache[T]
```

## Use the methods
```go
// create a cache
cache := docache.NewCache(ctx, interval, capacity, logger, doer)

// go do stuff on a goroutine
go cache.Loop()

// get a slice of the current data (parallel safe)
data := cache.Data()

// stop the loop when it's time to shutdown (blocks)
cache.Stop()

// start the cache looping again if you want
go cache.Loop()


// wait for multiple caches to stop before shutdown
var wg sync.WaitGroup

for _, cache := range caches {
    wg.Add(1)
    go func(cache docache.Cache){
        defer wg.Done()
        cache.Stop()
    }(cache)
}

wg.Wait()

// proceed to shutdown
```
