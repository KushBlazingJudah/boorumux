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
	"github.com/KushBlazingJudah/boorumux/filter"
)

var templates *template.Template
var indexRegexp = regexp.MustCompile(`^/([0-9a-z+]+)/?$`)
var mapPool = sync.Pool{
	New: func() any {
		return map[string]interface{}{}
	},
}
var ssPool = sync.Pool{
	New: func() any {
		return new([]string)
	},
}

type reqType uint

const (
	reqPost reqType = iota
	reqPage
	reqProxy
)

const (
	maxSidebarTags = 25
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

	// Blacklist is a list of blacklisted tags.
	// Posts containing these tags will not be shown in the page view, however
	// if explicitly requested they will be presented.
	Blacklist []filter.Filter

	boorus []string

	sync.Mutex
}

func init() {
	// Compile all of the templates.
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"embed":      func() error { panic("embed called too early") },
		"booruId":    func() error { panic("booruId called too early") },
		"unhumantag": func(s string) string { return strings.ReplaceAll(s, " ", "_") },
		"size":       humanSize,
		"pages":      buildPageBlock,
		"isUrl":      schemaRegexp.MatchString,
		"prettyUrl":  prettyUrl,
		"concat":     func(s []string, c string) string { return strings.Join(s, c) },
		"ver":        func() string { return "0.0.0" }, // TODO
	}).ParseGlob("./views/*.html"))
}

func checkin(d map[string]interface{}) {
	for k := range d {
		delete(d, k)
	}
	mapPool.Put(d)
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
	} else if ep == "/" {
		// Render the index, we don't need to do much for that though
		// Render it out
		tmpldata := mapPool.Get().(map[string]interface{})
		defer checkin(tmpldata)
		tmpldata["boorus"] = s.boorus

		templates.Funcs(template.FuncMap{"embed": func() error {
			return templates.Lookup("index.html").Execute(w, tmpldata)
		}}).ExecuteTemplate(w, "main.html", tmpldata)
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

	// Parse tags
	if t := r.URL.Query().Get("q"); t != "" {
		tags = strings.Split(t, " ")
	}

	switch action {
	case reqPage:
		s.pageHandler(w, r, targetBooru, v, tags)
	case reqPost:
		s.postHandler(w, r, targetBooru, v)
	case reqProxy:
		s.proxyHandler(w, r, targetBooru, r.URL.Query().Get("proxy"))
	}
}

func (s *Server) pageHandler(w http.ResponseWriter, r *http.Request, targetBooru string, page int, tags []string) {
	data, _, err := s.Boorus[targetBooru].Page(context.TODO(), booru.Query{Tags: tags}, page)
	if err != nil {
		panic(err)
	}

	// Filter out blacklisted tags
	// This looks weird but trust me on this; it's simply an in-place filter.
	n := 0
	for _, v := range data {
		fine := true
		for _, f := range s.Blacklist {
			if f.Match(&v) {
				fine = false
				break
			}
		}

		if fine {
			data[n] = v
			n++
		}
	}
	data = data[:n]

	// Find all of the tags on this page
	ssptr := ssPool.Get().(*[]string)
	defer ssPool.Put(ssptr)

	ss := *ssptr
	ss = ss[:0]

	for _, p := range data {
		ss = append(ss, p.Tags...)
	}

	pageTags := mostCommon(ss)
	if len(pageTags) > maxSidebarTags {
		pageTags = pageTags[:maxSidebarTags]
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

	// Render it out
	tmpldata := mapPool.Get().(map[string]interface{})
	defer checkin(tmpldata)

	tmpldata["booru"] = targetBooru
	tmpldata["boorus"] = s.boorus
	tmpldata["activeTags"] = tags
	tmpldata["tags"] = pageTags
	tmpldata["posts"] = data
	tmpldata["page"] = page
	tmpldata["q"] = r.URL.Query().Get("q")

	if tmpldata["q"] == "" {
		tmpldata["title"] = fmt.Sprintf("%s #%d - Boorumux", targetBooru, page)
	} else {
		tmpldata["title"] = fmt.Sprintf("%s #%d - Boorumux", tmpldata["q"], page)
	}

	templates.Funcs(template.FuncMap{
		"embed": func() error {
			return templates.Lookup("page.html").Execute(w, tmpldata)
		},
		"booruId": func(b booru.API) string {
			for k, v := range s.Boorus {
				if v == b {
					return k
				}
			}
			return ""
		},
	}).ExecuteTemplate(w, "main.html", tmpldata)
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request, targetBooru string, id int) {
	data, err := s.Boorus[targetBooru].Post(context.TODO(), id)
	if err != nil {
		panic(err)
	}

	// Make tags look nicer
	for i, v := range data.Tags {
		data.Tags[i] = strings.ReplaceAll(v, "_", " ")
	}

	// Sort it out
	sort.Strings(data.Tags)

	// Render it out
	tmpldata := mapPool.Get().(map[string]interface{})
	defer checkin(tmpldata)

	tmpldata["title"] = fmt.Sprintf("%s on %s - Boorumux", strings.Join(data.Tags, " "), targetBooru)
	tmpldata["booru"] = targetBooru
	tmpldata["boorus"] = s.boorus
	tmpldata["tags"] = data.Tags
	tmpldata["post"] = data
	tmpldata["q"] = r.URL.Query().Get("q")
	tmpldata["from"] = r.URL.Query().Get("from")

	templates.Funcs(template.FuncMap{"embed": func() error {
		return templates.Lookup("post.html").Execute(w, tmpldata)
	}}).ExecuteTemplate(w, "main.html", tmpldata)
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
}
