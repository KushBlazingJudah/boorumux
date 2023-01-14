package boorumux

import (
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

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

var mimeExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"video/mp4":  ".mp4",
}

func init() {
	// Compile all of the templates.
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"embed":      func() error { panic("embed called too early") },
		"booruId":    func() error { panic("booruId called too early") },
		"humantag":   func(s string) string { return strings.ReplaceAll(s, "_", " ") },
		"size":       humanSize,
		"pages":      buildPageBlock,
		"isUrl":      schemaRegexp.MatchString,
		"prettyUrl":  prettyUrl,
		"concat":     func(s []string, c string) string { return strings.Join(s, c) },
		"fmtTime":    func(t time.Time) string { return t.Format("2006-01-02 15:04:05 -0700") },
		"ver":        func() string { return verString },
		"mkUrl":      mkUrl,
		"has_string": has[string],
		"ext": func(m string) string {
			e, ok := mimeExt[m]
			if ok {
				return e
			}

			ex, _ := mime.ExtensionsByType(m)
			if ex == nil {
				return ""
			}
			return ex[0]
		},
	}).ParseGlob("./views/*.html"))
}

func checkin(d map[string]interface{}) {
	for k := range d {
		delete(d, k)
	}
	mapPool.Put(d)
}

func (s *Server) findBooru(r *http.Request, target string) (booru.API, error) {
	if target == "mux" {
		to, ok := r.URL.Query()["b"]
		if !ok {
			return nil, fmt.Errorf("b query parameter not found")
		}

		bs := make([]booru.API, len(to))
		for i, v := range to {
			b, ok := s.Boorus[v]
			if !ok {
				return nil, fmt.Errorf("booru \"%s\" not found", v)
			}

			bs[i] = b
		}

		return Mux(bs), nil
	}

	b, ok := s.Boorus[target]
	if !ok {
		return nil, fmt.Errorf("booru \"%s\" not found", target)
	}

	return b, nil
}

func (s *Server) pageHandler(w http.ResponseWriter, r *http.Request, targetBooru string, page int, tags []string) {
	tb, err := s.findBooru(r, targetBooru)
	if err != nil {
		panic(err)
	}

	data, _, err := tb.Page(r.Context(), booru.Query{Tags: tags}, page)
	if err != nil {
		panic(err)
	}

	reqTime := time.Now()

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

	if targetBooru == "mux" {
		tmpldata["mux"] = r.URL.Query()["b"]
	}

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

	t := template.Must(templates.Clone())
	t.Funcs(template.FuncMap{
		"embed": func() error {
			return t.Lookup("page.html").Execute(w, tmpldata)
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

	fmt.Fprintf(w, "<!-- rendered in %s -->", time.Since(reqTime).Truncate(time.Microsecond).String())
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request, targetBooru string, id int) {
	data, err := s.Boorus[targetBooru].Post(r.Context(), id)
	if err != nil {
		panic(err)
	}

	reqTime := time.Now()

	// Sort it out
	sort.Strings(data.Tags)

	// Render it out
	tmpldata := mapPool.Get().(map[string]interface{})
	defer checkin(tmpldata)

	if r.URL.Query().Get("from") == "mux" {
		tmpldata["mux"] = r.URL.Query()["b"]
	}

	tmpldata["title"] = fmt.Sprintf("%s on %s - Boorumux", strings.Join(data.Tags, " "), targetBooru)
	tmpldata["booru"] = targetBooru
	tmpldata["boorus"] = s.boorus
	tmpldata["tags"] = data.Tags
	tmpldata["post"] = data
	tmpldata["q"] = r.URL.Query().Get("q")
	tmpldata["from"] = r.URL.Query().Get("from")

	if s.Localbooru != "" {
		tmpldata["localbooru"] = s.Localbooru
	}

	t := template.Must(templates.Clone())
	t.Funcs(template.FuncMap{"embed": func() error {
		return t.Lookup("post.html").Execute(w, tmpldata)
	}}).ExecuteTemplate(w, "main.html", tmpldata)

	fmt.Fprintf(w, "<!-- rendered in %s -->", time.Since(reqTime).Truncate(time.Microsecond).String())
}

func (s *Server) saveHandler(w http.ResponseWriter, r *http.Request, targetBooru string, id int) {
	data, err := s.Boorus[targetBooru].Post(r.Context(), id)
	if err != nil {
		panic(err)
	}

	if err := s.save(r.Context(), data); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if rf := r.Referer(); rf != "" {
		// Redirect back
		w.Header().Set("Location", rf)
		w.WriteHeader(303) // See Other
		return
	}

	w.WriteHeader(201) // We did what we needed to do.
	// I would attempt to redirect back, but you don't allow me to.
}
