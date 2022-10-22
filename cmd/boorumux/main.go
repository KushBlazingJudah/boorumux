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
	"github.com/KushBlazingJudah/boorumux/filter"
)

var (
	Prefix = flag.String("prefix", "", "Root path of the server.")
	Listen = flag.String("addr", "localhost:8080", "Listening address of the HTTP server.")
)

type bcfg struct {
	Type, Url string
	Combine   []string
}

type cfg struct {
	Proxy     string
	Sources   map[string]bcfg
	Blacklist interface{}
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
	bm.Blacklist = filter.ParseMany(c.Blacklist)

	muxes := map[string]bcfg{}

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
		case "mux": // Dealt with later
			muxes[k] = v
		default:
			panic("unknown source")
		}
	}

	// Setup muxes
	for k, v := range muxes {
		m := boorumux.Mux{}

		// Check to see if all boorus are available
		for _, vv := range v.Combine {
			if b, ok := bm.Boorus[vv]; ok {
				m.Boorus = append(m.Boorus, b)
			} else {
				log.Fatal("for mux \"%s\": booru \"%s\" not found", k, vv)
			}
		}

		bm.Boorus[k] = &m
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
