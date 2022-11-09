package boorumux

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/KushBlazingJudah/boorumux/booru"
)

// Mux is the heart of the "-mux" suffix of Boorumux.
// It implements booru.API, but it also takes in anything that also implements
// booru.API.
type Mux struct {
	Boorus []booru.API
}

func (m *Mux) Page(ctx context.Context, q booru.Query, page int) ([]booru.Post, int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	var n int32 = 0
	var cerr error

	dc := make(chan []booru.Post, len(m.Boorus))

	for _, v := range m.Boorus {
		wg.Add(1)

		go func(b booru.API) {
			defer wg.Done()

			r, _, err := b.Page(ctx, q, page)
			if err != nil && !errors.Is(err, context.Canceled) {
				cerr = err
				cancel()
				return
			}

			if err == nil {
				dc <- r
				atomic.AddInt32(&n, int32(len(r)))
			}
		}(v)
	}

	wg.Wait()
	close(dc)

	if cerr != nil {
		return nil, -1, cerr
	}

	results := make([]booru.Post, 0, n)
	for v := range dc {
		results = append(results, v...)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Created.After(results[j].Created)
	})

	return results, -1, nil
}

func (m *Mux) Post(ctx context.Context, id int) (*booru.Post, error) {
	panic("tried to call Post on mux")
}

func (m *Mux) HTTP() *http.Client {
	return nil
}
