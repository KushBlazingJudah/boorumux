package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
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
	Prefix = flag.String("prefix", "", "Root path of the server. (unimplemented)")
	Listen = flag.String("addr", "localhost:8080", "Listening address of the HTTP server.")
)

var proxyFn func(*http.Request) (*url.URL, error) = nil

type bcfg struct {
	Type    string   `json:"type"`
	Proxy   string   `json:"proxy,omitempty"`
	Url     string   `json:"url,omitempty"`
	Combine []string `json:"combine,omitempty"`
}

type cfg struct {
	Proxy     string          `json:"proxy"`
	Sources   map[string]bcfg `json:"sources"`
	Blacklist interface{}     `json:"blacklist"`
}

func mkDefaults() {
	log.Printf("Creating ./boorumux.json with defaults")

	d := cfg{
		Sources: map[string]bcfg{
			"gelbooru": bcfg{
				Type: "gelbooru",
				Url:  "https://gelbooru.com",
			},
			"safebooru": bcfg{
				Type: "danbooru",
				Url:  "https://safebooru.donmai.us",
			},
			"mux": bcfg{
				Type:    "mux",
				Combine: []string{"gelbooru", "safebooru"},
			},
		},
		Blacklist: []string{
			// Safe selection of tags most may not want to see
			"guro",
			"scat",
			"furry",
			"loli",
		},
	}

	h, err := os.Create("./boorumux.json")
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	b, _ := json.MarshalIndent(d, "", "\t")
	if _, err := h.Write(b); err != nil {
		log.Fatalf("failed saving config: %v", err)
	}
}

func genBooru(name string, b bcfg, c cfg) booru.API {
	ht := &http.Client{}

	if b.Proxy != "" {
		pu, err := url.Parse(b.Proxy)
		if err != nil {
			log.Fatalf("error parsing proxy URL for booru \"%s\": %v", name, err)
		}

		log.Printf("booru \"%s\" proxy is %s", name, b.Proxy)
		ht.Transport = &http.Transport{Proxy: http.ProxyURL(pu)}
	} else {
		log.Printf("booru \"%s\" is using default proxy", name)
		ht.Transport = &http.Transport{Proxy: proxyFn}
	}

	u, err := url.Parse(b.Url)
	if err != nil {
		log.Fatalf("failed parsing url for booru \"%s\": %v", name, err)
	}

	var B booru.API
	switch b.Type {
	case "danbooru":
		B = &booru.Danbooru{
			URL:        u,
			HttpClient: ht,
		}
	case "gelbooru":
		B = &booru.Gelbooru{
			URL:        u,
			HttpClient: ht,
		}
	default:
		panic("unknown source")
	}

	return B
}

func main() {
	flag.Parse()

	bm := &boorumux.Server{
		Prefix: *Prefix,
	}

	f, err := os.Open("./boorumux.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			mkDefaults()
			f, err = os.Open("./boorumux.json")
			if err != nil {
				log.Fatalf("failed opening config: %v", err)
			}
		} else {
			log.Fatalf("failed opening config: %v", err)
		}
	}
	defer f.Close()

	c := cfg{}
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		log.Fatalf("failed reading config: %v", err)
	}

	// TODO: Config
	var pu *url.URL
	if c.Proxy != "" {
		pu, err = url.Parse(c.Proxy)
		if err != nil {
			log.Fatalf("error parsing proxy URL: %v", err)
		}

		log.Printf("Using global proxy %s", c.Proxy)
		proxyFn = http.ProxyURL(pu)
	}

	bm.Boorus = map[string]booru.API{}
	bm.Blacklist = filter.ParseMany(c.Blacklist)

	muxes := map[string]bcfg{}

	for k, v := range c.Sources {
		if v.Type == "mux" {
			muxes[k] = v
			continue
		}

		bm.Boorus[k] = genBooru(k, v, c)
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
