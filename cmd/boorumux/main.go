package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/KushBlazingJudah/boorumux"
	"github.com/KushBlazingJudah/boorumux/booru"
)

var (
	Prefix = flag.String("prefix", "", "Root path of the server.")
	Listen = flag.String("addr", "localhost:8080", "Listening address of the HTTP server.")
)

type cfg struct {
	Proxy   string
	Sources map[string]struct {
		Type, Url string
	}
	Blacklist []string
}

func main() {
	flag.Parse()

	bm := &boorumux.Server{
		Prefix: *Prefix,
	}
	f, err := os.Open("./boorumux.json")
	if err != nil {
		log.Fatalf("failed opening config: %v", err)
	}
	defer f.Close()

	c := cfg{}
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		log.Fatalf("failed reading config: %v", err)
	}

	// TODO: Config
	pu, _ := url.Parse("socks5://127.0.0.1:9050")
	ht := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(pu),
		},
	}

	bm.Boorus = map[string]booru.API{}

	for _, v := range c.Blacklist {
		if bm.Blacklist == nil {
			bm.Blacklist = make(map[string]struct{})
		}
		bm.Blacklist[v] = struct{}{}
	}

	for k, v := range c.Sources {
		u, err := url.Parse(v.Url)
		if err != nil {
			log.Fatalf("failed parsing url for %s: %v", k, err)
		}

		switch v.Type {
		case "danbooru":
			bm.Boorus[k] = &booru.Danbooru{
				URL:        u,
				HttpClient: ht,
			}
		case "gelbooru":
			bm.Boorus[k] = &booru.Gelbooru{
				URL:        u,
				HttpClient: ht,
			}
		default:
			panic("unknown source")
		}
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
