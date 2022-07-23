package boorumux

import (
	"context"
	"html/template"
	"net/http"
	"io"
	"fmt"
	"sync"
	"strings"
	"sort"

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
	// This should be filled in from a config file, and not written to after
	// the server starts.
	Boorus map[string]booru.API

	boorus []string

	sync.Mutex
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

	if len(s.boorus) != len(s.Boorus) {
		// Need to regenerate the text-only booru list
		s.Lock()
		s.boorus = make([]string, 0, len(s.Boorus))
		for k := range s.Boorus {
			s.boorus = append(s.boorus, k)
		}
		s.Unlock()
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

	// Find all of the tags on this page
	// To keep things simple, we're going to simply use a map[string]struct{}.
	// This ensures that results are unique, we just need to convert it into a
	// string slice.
	tagmap := map[string]struct{}{}
	for _, p := range test {
		for _, v := range p.Tags {
			tagmap[v] = struct{}{}
		}
	}
	tags := make([]string, 0, len(tagmap))
	for k := range tagmap {
		tags = append(tags, strings.ReplaceAll(k, "_", " "))
	}
	sort.Strings(tags)

	templates.ExecuteTemplate(w, "page.html", map[string]interface{}{
		"booru": "test",
		"boorus": s.boorus,
		"tags": tags,
		"posts": test,
	})
}
