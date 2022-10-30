package boorumux

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/KushBlazingJudah/boorumux/booru"
)

var templates *template.Template
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

func init() {
	// Compile all of the templates.
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"embed":     func() error { panic("embed called too early") },
		"booruId":   func() error { panic("booruId called too early") },
		"humantag":  func(s string) string { return strings.ReplaceAll(s, "_", " ") },
		"size":      humanSize,
		"pages":     buildPageBlock,
		"isUrl":     schemaRegexp.MatchString,
		"prettyUrl": prettyUrl,
		"concat":    func(s []string, c string) string { return strings.Join(s, c) },
		"ver":       func() string { return verString }, // TODO
	}).ParseGlob("./views/*.html"))
}

func checkin(d map[string]interface{}) {
	for k := range d {
		delete(d, k)
	}
	mapPool.Put(d)
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
		for _, v := range p.Tags {
			ok := true

			for _, k := range tags {
				if v == k {
					ok = false
					break
				}
			}

			if ok {
				ss = append(ss, v)
			}
		}
	}

	pageTags := mostCommon(ss)
	if len(pageTags) > maxSidebarTags {
		pageTags = pageTags[:maxSidebarTags]
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
