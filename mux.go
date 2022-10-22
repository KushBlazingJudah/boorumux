package boorumux

import (
	"context"
	"net/http"

	"github.com/KushBlazingJudah/boorumux/booru"
)

// Mux is the heart of the "-mux" suffix of Boorumux.
// It implements booru.API, but it also takes in anything that also implements
// booru.API.
type Mux struct {
	Boorus []booru.API
}

func (m *Mux) Page(ctx context.Context, q booru.Query, page int) ([]booru.Post, int, error) {
	// TODO: Send multiple requests at once.

	results := []booru.Post{}

	for _, v := range m.Boorus {
		r, _, err := v.Page(ctx, q, page)
		if err != nil {
			return results, -1, err
		}

		results = append(results, r...)
	}

	return results, -1, nil
}

func (m *Mux) Post(ctx context.Context, id int) (*booru.Post, error) {
	panic("tried to call Post on mux")
}

func (m *Mux) HTTP() *http.Client {
	return nil
}
