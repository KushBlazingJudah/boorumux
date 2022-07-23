package boorumux

import (
	"context"
	"html/template"
	"net/http"
	"io"
	"fmt"

	"github.com/KushBlazingJudah/boorumux/booru"
)

var templates *template.Template
var test []booru.Post

// Server holds the main configuration for Boorumux and doubles as a
// http.Handler.
// The zero-value is usable.
type Server struct {
	// Prefix is the root directory of this server.
	// Responses will be relative to this directory.
	// Strip the prefix on incoming requests according to this value.
	Prefix string

	// Boorus is a mapping of human readable names to booru APIs.
	// This should be filled in from a config file.
	Boorus map[string]booru.API
}

func init() {
	// Compile all of the templates.
	templates = template.Must(template.New("").ParseGlob("./views/*.html"))
}

// ServeHTTP serves a requested page.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		// We only support GET
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Call upon test booru for test data
	if test == nil {
		var err error
		test, err = s.Boorus["test"].Page(context.TODO(), booru.Query{}, 0)
		if err != nil {
			panic(err)
		}
	}

	ep := r.URL.EscapedPath()
	if ep == "/proxy" {
		// We're proxying a page!
		target := r.URL.Query().Get("t")

		// TODO: Trusted domains
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			panic(err)
		}

		res, err := s.Boorus["test"].HTTP().Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		w.Header().Set("Content-Length", fmt.Sprint(res.ContentLength))
		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		if _, err := io.Copy(w, res.Body); err != nil {
			panic(err)
		}
	}

	templates.ExecuteTemplate(w, "page.html", test)
}
