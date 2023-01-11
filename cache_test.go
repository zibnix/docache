package docache

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

type Fibber struct {
	p1, p2 int
	i      int
}

type fib struct {
	val   int
	index int
}

func (f *Fibber) Do() (*fib, error) {
	n := f.p1 + f.p2
	f.p1 = f.p2
	f.p2 = n
	f.i++

	return &fib{
		val:   f.p2,
		index: f.i,
	}, nil
}

// Fibber implements Doer[*fib] and emits the fibonacci sequence
func NewFibber() *Fibber {
	return &Fibber{
		p1: 0,
		p2: 1,
		i:  2,
	}
}

// errs implements Doer[T], and returns a T
// on the first call to Do, but only errors after that
type errs[T any] struct {
	c      int
	getter func() T
}

func (e *errs[T]) Do() (T, error) {
	defer func() {
		e.c++
	}()

	if e.c == 0 {
		return e.getter(), nil
	}

	return *new(T), fmt.Errorf("#%d", e.c)
}

var (
	ctx      = context.Background()
	interval = 10 * time.Millisecond
	capacity = 100
	logger   = log.New(io.Discard, "test logger: ", log.Lshortfile)
	sleep    = 50 * time.Millisecond
)

func newCache(capacity int) Cache[*fib] {
	return NewCache[*fib](ctx, interval, capacity, logger, NewFibber())
}

func TestLatest(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	e := &errs[int64]{
		getter: func() int64 {
			return r.Int63()
		},
	}

	cache := NewCache[int64](ctx, interval, capacity, logger, e)
	go cache.Loop()

	time.Sleep(sleep)

	latest := cache.Latest()

	time.Sleep(sleep)

	if !reflect.DeepEqual(latest, cache.Latest()) {
		t.Fatal("latest value should not change between calls with an errs Doer")
	}
}

func TestStop(t *testing.T) {
	t.Parallel()
	cacheStartStop(t, newCache(capacity))
}

func cacheStartStop(t *testing.T, cache Cache[*fib]) {
	looping := make(chan struct{})
	go func() {
		cache.Loop()
		close(looping)
	}()

	time.Sleep(sleep)

	stop := make(chan struct{})
	go func() {
		cache.Stop()
		close(stop)
	}()

	timer := time.NewTimer(sleep)
	select {
	case <-timer.C:
		t.Fatal("timeout while waiting for Stop to return")
	case <-stop:
		// success
	}

	timer = time.NewTimer(sleep)
	select {
	case <-timer.C:
		t.Fatal("timeout while waiting for loop to return after Stop")
	case <-looping:
		// success
	}
}

func TestData(t *testing.T) {
	t.Parallel()
	cache := newCache(capacity)

	cacheStartStop(t, cache)

	d1 := cache.Data()
	validateData(t, d1)

	cacheStartStop(t, cache)

	d2 := cache.Data()
	validateData(t, d2)

	if reflect.DeepEqual(d1, d2) {
		t.Fatal("raw cache data was deep equal after multiple starts and stops")
	}
}

func validateData(t *testing.T, data []Data[*fib]) {
	if len(data) <= 0 {
		t.Fatal("no data generated")
	}

	if len(data) < 3 {
		t.Fatalf("less than 3 data generated in %v", sleep)
	}

	var prev Data[*fib]
	for i, d := range data {
		if i != 0 {
			if d.Error != nil {
				t.Fatalf("unexpected error: %+v", prev.Error)
			}

			if prev.Timestamp.After(d.Timestamp) {
				t.Fatal("data out of order")
			}

			if prev.Value.index == d.Value.index {
				t.Fatal("index did not change between data")
			}

			if prev.Value.val == d.Value.val {
				t.Fatal("value did not change between data")
			}
		}

		prev = d
	}
}

func TestDataOverflow(t *testing.T) {
	t.Parallel()
	capacity := 5
	cache := newCache(capacity)

	go cache.Loop()

	time.Sleep(2 * sleep)

	cache.Stop()

	latest := cache.Latest()

	if latest.Value.index < capacity+2 {
		t.Fatal("latest index did not exceed capacity")
	}

	slc := cache.Data()
	if len(slc) != capacity {
		t.Fatal("capacity not reached or not respected")
	}

	if !reflect.DeepEqual(latest, slc[len(slc)-1]) {
		t.Fatal("data not in correct state after capacity overflow")
	}

	validateData(t, slc)
}

func TestChaos(t *testing.T) {
	t.Parallel()
	cache := newCache(capacity)

	quit := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				timer := time.NewTimer(interval / 2)
				select {
				case <-timer.C:
					callMethod[*fib](cache, r.Intn(4))
				case <-quit:
					return
				}
			}
		}()
	}

	time.Sleep(2 * sleep)
	close(quit)
	wg.Wait()
}

func callMethod[T any](cache Cache[T], m int) {
	switch m {
	case 0:
		go cache.Loop()
	case 1:
		cache.Stop()
	case 2:
		cache.Latest()
	case 3:
		cache.Data()
	}
}
