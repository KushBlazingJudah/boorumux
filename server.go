package boorumux

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/KushBlazingJudah/boorumux/booru"
)

var templates *template.Template
var indexRegexp = regexp.MustCompile(`^/([0-9a-z+]+)/?$`)

type reqType uint

const (
	reqPost reqType = iota
	reqPage
	reqProxy
)

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

	ep := r.URL.EscapedPath()
	if ep == "/favicon.ico" {
		return
	}

	// Determine what kind of request this is
	// Usually it'll be a page/post request, look for that
	targetBooru := ""
	action := reqPage
	v := 0
	var tags []string
	var err error

	// Check if it matches the index regexp
	matches := indexRegexp.FindStringSubmatch(ep)
	if len(matches) > 0 {
		targetBooru = matches[1]
	} else {
		// TODO
		panic("invalid request")
	}

	// Determine action
	if p := r.URL.Query().Get("page"); p != "" {
		// Page request
		v, err = strconv.Atoi(p)
		if err != nil {
			// TODO
			panic(err)
		}
		action = reqPage

		// Parse tags
		if t := r.URL.Query().Get("tags"); t != "" {
			tags = strings.Split(t, " ")
			fmt.Println(tags) // DEBUG
		}
	} else if p := r.URL.Query().Get("post"); p != "" {
		// Page request
		v, err = strconv.Atoi(p)
		if err != nil {
			// TODO
			panic(err)
		}
		action = reqPost
	} else if r.URL.Query().Get("proxy") != "" {
		action = reqProxy
	}

	switch action {
	case reqPage:
		s.pageHandler(w, r, targetBooru, v, tags)
	case reqPost:
		s.pageHandler(w, r, targetBooru, v, nil)
	case reqProxy:
		s.proxyHandler(w, r, targetBooru, r.URL.Query().Get("proxy"))
	}
}

func (s *Server) pageHandler(w http.ResponseWriter, r *http.Request, targetBooru string, page int, tags []string) {
	fmt.Println(targetBooru)
	data, err := s.Boorus[targetBooru].Page(context.TODO(), booru.Query{Tags: tags}, page)
	if err != nil {
		panic(err)
	}

	// Find all of the tags on this page
	// To keep things simple, we're going to simply use a map[string]struct{}.
	// This ensures that results are unique, we just need to convert it into a
	// string slice.
	tagmap := map[string]struct{}{}
	for _, p := range data {
		for _, v := range p.Tags {
			tagmap[v] = struct{}{}
		}
	}

	pageTags := make([]string, 0, len(tagmap))
	for k := range tagmap {
		pageTags = append(pageTags, k)
	}

	// Remove tags from pageTags
	for _, v := range tags {
		for i, x := range pageTags {
			if x == v {
				pageTags[i] = pageTags[len(pageTags)-1]
				pageTags = pageTags[:len(pageTags)-1]
				break
			}
		}
	}

	// Make things look nicer
	for i, v := range pageTags {
		pageTags[i] = strings.ReplaceAll(v, "_", " ")
	}
	for i, v := range tags {
		tags[i] = strings.ReplaceAll(v, "_", " ")
	}

	// Sort it out
	sort.Strings(pageTags)
	sort.Strings(tags)

	// Render it out
	templates.ExecuteTemplate(w, "page.html", map[string]interface{}{
		"booru":      targetBooru,
		"boorus":     s.boorus,
		"activeTags": tags,
		"tags":       pageTags,
		"posts":      data,
		"page":       page,
	})
}

func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request, targetBooru string, target string) {
	// TODO: Trusted domains
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		panic(err)
	}

	res, err := s.Boorus[targetBooru].HTTP().Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	w.Header().Set("Content-Length", fmt.Sprint(res.ContentLength))
	w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	if _, err := io.Copy(w, res.Body); err != nil {
		panic(err)
	}
	return
}