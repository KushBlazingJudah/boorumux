package boorumux

import (
	"fmt"
	"sort"
	"sync"
)

// counter can count items according to how often they appear.
//
// counter is not thread safe.
type counter map[interface{}]int

var counterPool = sync.Pool{
	New: func() any {
		return counter{}
	},
}

func (c counter) count(a interface{}) {
	n, ok := c[a]
	if !ok {
		c[a] = 1
	} else {
		c[a] = n + 1
	}
}

func (c counter) reset() {
	for k := range c {
		delete(c, k)
	}
}

func humanSize(b int) string {
	var f float64
	var r string
	if b > 1024*1024 {
		f = float64(b) / (1024 * 1024)
		r = "MB"
	} else if b > 1024 {
		f = float64(b) / 1024
		r = "KB"
	} else {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.2f %s", f, r)
}

// mostCommon returns the most common items in a list.
func mostCommon[T any](list []T) []T {
	c := counterPool.Get().(counter)

	// defers are LIFO; this resets it first and then checks it into the pool
	defer counterPool.Put(c)
	defer c.reset()

	for _, v := range list {
		c.count(v)
	}

	o := make([]T, 0, len(c))
	for k := range c {
		o = append(o, k.(T))
	}

	sort.Slice(o, func(i, j int) bool {
		return c[i] < c[j]
	})

	return o
}
