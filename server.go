package boorumux

import (
	"html/template"
	"net/http"

	"github.com/KushBlazingJudah/boorumux/booru"
)

var templates *template.Template

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

	_ = r.URL.EscapedPath()
	templates.ExecuteTemplate(w, "page.html", nil)
}
