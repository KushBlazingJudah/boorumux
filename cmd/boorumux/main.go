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

type cfg struct {
	Proxy      string                 `json:"proxy"`
	Sources    map[string]interface{} `json:"sources"`
	Localbooru string                 `json:"localbooru"`
	Blacklist  interface{}            `json:"blacklist"`
}

func mkDefaults() {
	log.Printf("Creating ./boorumux.json with defaults")

	d := cfg{
		Sources: map[string]interface{}{
			"gelbooru": map[string]string{
				"type": "gelbooru",
				"url":  "https://gelbooru.com",
			},
			"safebooru": map[string]string{
				"type": "danbooru",
				"url":  "https://safebooru.donmai.us",
			},
			"examplemux": map[string]interface{}{
				"type":    "mux",
				"combine": []string{"gelbooru", "safebooru"},
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

func genBooru(name string, b map[string]interface{}, c cfg) booru.API {
	ht := &http.Client{}

	if b["proxy"] != nil {
		if b["proxy"].(string) != "none!" {
			pu, err := url.Parse(b["proxy"].(string))
			if err != nil {
				log.Fatalf("error parsing proxy URL for booru \"%s\": %v", name, err)
			}

			log.Printf("booru \"%s\" proxy is %s", name, pu.String())
			ht.Transport = &http.Transport{Proxy: http.ProxyURL(pu)}
		} else {
			log.Printf("booru \"%s\" is using *no* proxy", name)
		}
	} else {
		log.Printf("booru \"%s\" is using default proxy", name)
		ht.Transport = &http.Transport{Proxy: proxyFn}
	}

	// Default arguments
	if _, ok := b["agent"]; !ok {
		b["agent"] = boorumux.UserAgent
	}

	b["http"] = ht

	B, err := booru.New(b["type"].(string), b)
	if err != nil {
		log.Fatalf("error initializing booru \"%s\": %v", name, err)
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

	bm.Localbooru = c.Localbooru

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

	muxes := map[string]map[string]interface{}{}

	for k, v := range c.Sources {
		vv := v.(map[string]interface{})
		if vv["type"] == "mux" {
			muxes[k] = vv
			continue
		}

		bm.Boorus[k] = genBooru(k, vv, c)
	}

	// Setup muxes
	for k, v := range muxes {
		m := boorumux.Mux{}

		// Check to see if all boorus are available
		for _, vv := range v["combine"].([]interface{}) {
			if b, ok := bm.Boorus[vv.(string)]; ok {
				m = append(m, b)
			} else {
				log.Fatal("for mux \"%s\": booru \"%s\" not found", k, vv)
			}
		}

		bm.Boorus[k] = m
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
