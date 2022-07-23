package booru

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Gelbooru implements the Gelbooru API.
type Gelbooru struct {
	// URL is the location of where the Gelbooru API is.
	// This is a necessary field, or else all requests will fail as they have
	// nowhere to go.
	// URL must not change after first use.
	//
	// Example: "https://gelbooru.com"
	URL *url.URL

	// HttpClient is the HTTP client object that is used to talk to the
	// Gelbooru API.
	HttpClient *http.Client
}

// gelbooruPost holds some of the information returned by the Gelbooru API.
// This isn't supposed to be used outside of this package; it is simply here to
// ease unmarshaling of responses.
// Always convert to the standard Post struct instead.
type gelbooruPost struct {
	Id int

	Created string `json:"created_at"`
	Updated int64  `json:"change"`

	Score int

	Source string

	OriginalUrl string `json:"file_url"`
	ThumbUrl    string `json:"preview_url"`

	Tags string `json:"tags"`

	// Rating
}

type gelbooruResp struct {
	Post []gelbooruPost
}

// toPost converts the internal representation to an actual Post used by the
// outer world.
func (dp gelbooruPost) toPost() Post {
	// Some things are 1:1 but others need to be parsed
	p := Post{
		Id:     dp.Id,
		Score:  dp.Score,
		Source: dp.Source,
		Tags:   strings.Split(dp.Tags, " "),
		Images: []Image{
			{
				Href:      dp.OriginalUrl,
				MIME:      mime.TypeByExtension(filepath.Ext(dp.OriginalUrl)), // mime asks we include the dot
				Size:      0,                                                  // never told
				Thumbnail: false,
			},
			{
				Href:      dp.ThumbUrl,
				MIME:      "image/jpeg", // assumption
				Size:      0,            // we are never told
				Thumbnail: true,
			},
		},
	}

	p.Created, _ = time.Parse(time.RubyDate, dp.Created)
	p.Updated = time.Unix(dp.Updated, 0)

	return p
}

// HTTP returns the HttpClient that this booru uses.
func (d *Gelbooru) HTTP() *http.Client {
	return d.HttpClient
}

func (d *Gelbooru) Page(ctx context.Context, q Query, page int) ([]Post, error) {
	urlq := queryify(map[string]string{
		"page": "dapi",
		"s":    "post",
		"q":    "index",
		"pid":  fmt.Sprint(page),
		"tags": strings.Join(q.Tags, " "),
		"json": "1",
	})

	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = filepath.Join(u.Path, "/index.php")

	if u.RawQuery != "" {
		// Something is already here
		u.RawQuery += "&"
	}

	u.RawQuery += urlq

	// Create a request object
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Do the needful
	res, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Something bad happened, ditch
		return nil, fmt.Errorf("gelbooru: unknown error (%d)", res.StatusCode)
	}

	// Parse the results
	var rawResp gelbooruResp

	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, err
	}

	// Convert
	out := make([]Post, len(rawResp.Post))
	for i, v := range rawResp.Post {
		out[i] = v.toPost()
	}

	return out, nil
}

func (d *Gelbooru) Post(ctx context.Context, id int) (*Post, error) {
	urlq := queryify(map[string]string{
		"page": "dapi",
		"s":    "post",
		"q":    "index",
		"id":   fmt.Sprint(id),
		"json": "1",
	})

	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = filepath.Join(u.Path, "/index.php")

	if u.RawQuery != "" {
		// Something is already here
		u.RawQuery += "&"
	}

	u.RawQuery += urlq

	// Create a request object
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Do the needful
	res, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Something bad happened, ditch
		return nil, fmt.Errorf("gelbooru: unknown error (%d)", res.StatusCode)
	}

	// Parse the results
	var rawResp gelbooruResp
	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, err
	}

	// Convert
	if len(rawResp.Post) == 0 {
		return nil, fmt.Errorf("gelbooru: not found")
	}

	out := rawResp.Post[0].toPost()
	return &out, nil
}
