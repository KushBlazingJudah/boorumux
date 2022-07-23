package main

import (
	"flag"
	"net/http"
	"net/url"
	"time"

	"github.com/KushBlazingJudah/boorumux"
	"github.com/KushBlazingJudah/boorumux/booru"
)

var (
	Prefix = flag.String("prefix", "", "Root path of the server.")
	Listen = flag.String("addr", "localhost:8080", "Listening address of the HTTP server.")
)

func main() {
	flag.Parse()

	bm := &boorumux.Server{
		Prefix: *Prefix,
	}
	db, _ := url.Parse("https://safebooru.donmai.us")
	pu, _ := url.Parse("socks5://127.0.0.1:9050")
	bm.Boorus = map[string]booru.API{
		"test": &booru.Danbooru{
			URL: db,
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(pu),
				},
			},
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./static/css"))))
	mux.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./static/js"))))
	mux.Handle("/", bm)

	// WriteTimeout remains commented to allow us to send large files to the client.
	s := http.Server{
		Addr:        *Listen,
		Handler:     mux,
		ReadTimeout: time.Second * 15,
		// WriteTimeout: time.Second*10,
		MaxHeaderBytes: 1 << 20,
		IdleTimeout:    time.Minute * 2,
	}

	// TODO: StripPrefix
	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
