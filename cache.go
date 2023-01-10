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
	logger   *log.Logger
	doer     Doer[T]

	latest Data[T]
	data   []Data[T]

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

func (c *cache[T]) Latest() Data[T] {
	c.lk.RLock()
	defer c.lk.RUnlock()
	return c.latest
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

func (c *cache[T]) finished() {
	c.lk.Lock()
	c.looping = false
	c.lk.Unlock()
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

func (c *cache[T]) add(d Data[T]) {
	c.lk.Lock()
	if d.Error == nil {
		c.latest = d
	}

	if c.capacity > 0 {
		c.grow(d)
	}
	c.lk.Unlock()
}

// an alternative version of grow is included below
// call with c.lk.Lock held
func (c *cache[T]) grow(d Data[T]) {
	if len(c.data) < c.capacity {
		c.data = append(c.data, d)
		return
	}

	// we shift values in-place every time, rather than allow append()
	// to grow the backing array and copy as it sees fit
	for i := 0; i < len(c.data)-1; i++ {
		c.data[i] = c.data[i+1]
	}

	c.data[len(c.data)-1] = d
}

// UNUSED:
// an alternative implementation of grow that always appends
// and reslices when capacity is exceeded
// call with c.lk.Lock held
func (c *cache[T]) _grow(d Data[T]) {
	c.data = append(c.data, d)

	length := len(c.data)
	if length <= c.capacity {
		return
	}

	overflow := length - c.capacity

	// zero out overflow of backing array to encourage GC
	for i := 0; i < overflow; i++ {
		c.data[i] = Data[T]{}

	}

	c.data = c.data[overflow:]
}
