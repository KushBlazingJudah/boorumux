package boorumux

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/KushBlazingJudah/boorumux/booru"
	"github.com/KushBlazingJudah/boorumux/filter"
)

type reqType uint

const (
	reqPost reqType = iota
	reqPage
	reqProxy
)

const (
	maxSidebarTags = 25
)

var indexRegexp = regexp.MustCompile(`^/([0-9a-z+]+)/?$`)
var proxyRegexp = regexp.MustCompile(`^/([0-9a-z+]+)/proxy/[^/]*`)

// proxyReqHeaders is a list of headers that are sent with a proxy request to a
// booru.
var proxyReqHeaders = []string{
	"Range",
}

// proxyRespHeaders is a list of headers copied from the response from the
// server for a proxy request.
var proxyRespHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Accept-Ranges",
	"Last-Modified",
	"Etag",
	"Expires",
	"Cache-Control",
}

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

// ServeHTTP serves a requested page.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", serverHeader)

	if r.Method != "GET" && r.Method != "HEAD" {
		// We only support GET and HEAD
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
		w.WriteHeader(http.StatusNotFound)
		return
	} else if ep == "/" {
		// Render the index, we don't need to do much for that though
		// Render it out
		tmpldata := mapPool.Get().(map[string]interface{})
		defer checkin(tmpldata)
		tmpldata["boorus"] = s.boorus

		t := template.Must(templates.Clone())
		t.Funcs(template.FuncMap{"embed": func() error {
			return t.Lookup("index.html").Execute(w, tmpldata)
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
		// Does it match the proxy regexp?
		matches = proxyRegexp.FindStringSubmatch(ep)
		if len(matches) > 0 {
			// Yes it does!
			s.proxyHandler(w, r, matches[1], r.URL.Query().Get("proxy"))
			return
		}

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
	}
}

func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request, targetBooru string, target string) {
	defer func() {
		// XXX: There's this really annoying bug and I'm not sure if it's our
		// fault or Go's, but essentially under some circumstances io.Copy will
		// panic with a slice out of bounds error, something along those lines.
		// I forget.
		//
		// Point is, I don't know what actually causes the problem.
		// I suspect it is a race condition.
		// If you are a brave soul, remove this deferred function and trigger
		// it somehow, because I also don't know what triggers it.
		if v := recover(); v != nil {
			log.Printf("proxyHandler panic: %v", v)
		}
	}()

	// TODO: Trusted domains
	req, err := http.NewRequest(r.Method, target, nil)
	if err != nil {
		panic(err)
	}

	// Copy over request headers if they're there
	for _, k := range proxyReqHeaders {
		if v := r.Header.Get(k); v != "" {
			req.Header.Set(k, v)
		}
	}

	res, err := s.Boorus[targetBooru].HTTP().Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Copy over some headers if they're there
	for _, k := range proxyRespHeaders {
		if v := res.Header.Get(k); v != "" {
			w.Header().Set(k, v)
		}
	}

	if r.Method != "HEAD" { // We don't send a message body for HEAD
		// This is where the aforementioned bug occurs.
		if _, err := io.Copy(w, res.Body); err != nil {
			panic(err)
		}
	}
}
