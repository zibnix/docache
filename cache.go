package docache

import (
	"context"
	"log"
	"sync"
	"time"
)

type cache[T any] struct {
	ctx      context.Context
	interval time.Duration
	capacity int
	logger   log.Logger
	doer     Doer[T]

	data []Data[T]

	looping bool
	lk      sync.RWMutex
	wg      sync.WaitGroup
	quit    chan struct{}
}

func (c *cache[T]) Data() []Data[T] {
	c.lk.RLock()
	cpy := make([]Data[T], 0, len(c.data))

	for _, d := range c.data {
		cpy = append(cpy, d)
	}
	c.lk.RUnlock()

	return cpy
}

func (c *cache[T]) Loop() {
	c.lk.Lock()
	if c.looping {
		c.lk.Unlock()
		return
	}
	c.looping = true
	c.quit = make(chan struct{})
	c.wg.Add(1)
	defer c.wg.Done()
	c.lk.Unlock()

	c.add(c.do())

	for {
		timer := time.NewTimer(c.interval)

		select {
		case <-timer.C:
			c.add(c.do())
		case <-c.ctx.Done():
			c.finished()
			return
		case <-c.quit:
			c.finished()
			return
		}
	}
}

func (c *cache[T]) do() Data[T] {
	ts := time.Now()
	v, err := c.doer.Do()
	if err != nil {
		c.logger.Printf("Doer error: %+v\n", err)
	} else {
		c.logger.Printf("Doer success: %+v\n", v)
	}

	return Data[T]{
		Value:     v,
		Timestamp: ts,
		Error:     err,
	}
}

func (c *cache[T]) Stop() {
	c.lk.Lock()
	if !c.looping {
		c.lk.Unlock()
		return
	}

	close(c.quit)
	c.lk.Unlock()

	c.wg.Wait()
}

// an alternative version of add is included at the bottom of this file
func (c *cache[T]) add(d Data[T]) {
	c.lk.Lock()
	if len(c.data) < c.capacity {
		c.data = append(c.data, d)
		c.lk.Unlock()
		return
	}

	// we shift values in-place every time, rather than allow append()
	// to grow the backing array and copy as it sees fit
	for i := 0; i < len(c.data)-1; i++ {
		c.data[i] = c.data[i+1]
	}

	c.data[len(c.data)-1] = d
	c.lk.Unlock()
}

func (c *cache[T]) finished() {
	c.lk.Lock()
	c.looping = false
	c.lk.Unlock()
}

// UNUSED:
// an alternative implementation of add that always appends
// and reslices when capacity is exceeded
func (c *cache[T]) _add(d Data[T]) {
	c.lk.Lock()
	c.data = append(c.data, d)

	length := len(c.data)
	if length <= c.capacity {
		c.lk.Unlock()
		return
	}

	overflow := length - c.capacity

	// zero out overflow of backing array to encourage GC
	for i := 0; i < overflow; i++ {
		// new() allocates any type and returns a pointer to it
		// dereferencing that yields the zero value of a T
		c.data[i] = Data[T]{Value: *new(T)}

	}

	c.data = c.data[overflow:]
	c.lk.Unlock()
}
