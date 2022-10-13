package boorumux

import (
	"context"
	"net/http"

	"github.com/KushBlazingJudah/boorumux/booru"
)

// mux is the heart of the "-mux" suffix of Boorumux.
// It implements booru.API, but it also takes in anything that also implements
// booru.API.
type mux struct {
	Boorus []booru.API
}

func (m *mux) Page(ctx context.Context, q booru.Query, page int) ([]booru.Post, int, error) {
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

func (m *mux) Post(ctx context.Context, id int) (*booru.Post, error) {
	panic("tried to call Post on mux")
}

func (m *mux) HTTP() *http.Client {
	return nil
}
