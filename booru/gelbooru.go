package booru

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// gelbooru implements the gelbooru API.
//
// API documentation: https://gelbooru.com/index.php?page=wiki&s=view&id=18780
type gelbooru struct {
	// URL is the location of where the gelbooru API is.
	// This is a necessary field, or else all requests will fail as they have
	// nowhere to go.
	// URL must not change after first use.
	//
	// Example: "https://gelbooru.com"
	URL *url.URL

	// HttpClient is the HTTP client object that is used to talk to the
	// gelbooru API.
	HttpClient *http.Client

	ua string
}

// gelbooruPost holds some of the information returned by the gelbooru API.
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

	Width, Height int
	PreviewWidth  int `json:"preview_width"`
	PreviewHeight int `json:"preview_height"`

	Rating string
}

type gelbooruResp struct {
	A struct {
		Limit, Offset, Total int
	} `json:"@attributes"`
	Post []gelbooruPost
}

func init() {
	registered["gelbooru"] = func(cfg map[string]interface{}) (API, error) {
		g := &gelbooru{}

		g.ua = cfg["agent"].(string)
		g.HttpClient = cfg["http"].(*http.Client)

		u, err := url.Parse(cfg["url"].(string))
		if err != nil {
			return nil, fmt.Errorf("failed parsing url: %w", err)
		}

		g.URL = u

		return g, nil
	}
}

// toPost converts the internal representation to an actual Post used by the
// outer world.
func (dp gelbooruPost) toPost(d *gelbooru) Post {
	// Some things are 1:1 but others need to be parsed
	p := Post{
		Id:     dp.Id,
		Score:  dp.Score,
		Source: dp.Source,
		Tags:   strings.Split(dp.Tags, " "),
		Original: Image{
			Href:   dp.OriginalUrl,
			MIME:   mime.TypeByExtension(path.Ext(dp.OriginalUrl)), // mime asks we include the dot
			Size:   0,                                              // never told
			Width:  dp.Width,
			Height: dp.Height,
		},
		Thumbnail: Image{
			Href:   dp.ThumbUrl,
			MIME:   "image/jpeg", // assumption
			Size:   0,            // we are never told
			Width:  dp.PreviewWidth,
			Height: dp.PreviewHeight,
		},
		Origin: d,
	}

	switch dp.Rating {
	default:
		fallthrough
	case "general":
		p.Rating = General
	case "questionable":
		p.Rating = Questionable
	case "sensitive":
		p.Rating = Sensitive
	case "explicit":
		p.Rating = Explicit
	}

	p.Created, _ = time.Parse(time.RubyDate, dp.Created)
	p.Updated = time.Unix(dp.Updated, 0)

	return p
}

// HTTP returns the HttpClient that this booru uses.
func (d *gelbooru) HTTP() *http.Client {
	return d.HttpClient
}

func (d *gelbooru) Page(ctx context.Context, q Query, page int) ([]Post, int, error) {
	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = path.Join(u.Path, "/index.php")
	uq := u.Query()
	uq.Set("page", "dapi")
	uq.Set("s", "post")
	uq.Set("q", "index")
	uq.Set("pid", fmt.Sprint(page))
	uq.Set("tags", strings.Join(q.Tags, " "))
	uq.Set("json", "1")
	u.RawQuery = uq.Encode()

	// Create a request object
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", d.ua)

	// Do the needful
	res, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Something bad happened, ditch
		return nil, 0, newHTTPError(res)
	}

	// Parse the results
	var rawResp gelbooruResp

	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, 0, err
	}

	// Convert
	out := make([]Post, len(rawResp.Post))
	for i, v := range rawResp.Post {
		out[i] = v.toPost(d)
	}

	return out, int(math.Ceil(float64(rawResp.A.Total-rawResp.A.Offset) / float64(rawResp.A.Limit))), nil
}

func (d *gelbooru) Post(ctx context.Context, id int) (*Post, error) {
	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = path.Join(u.Path, "/index.php")
	uq := u.Query()
	uq.Set("page", "dapi")
	uq.Set("s", "post")
	uq.Set("q", "index")
	uq.Set("id", fmt.Sprint(id))
	uq.Set("json", "1")
	u.RawQuery = uq.Encode()

	// Create a request object
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", d.ua)

	// Do the needful
	res, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Something bad happened, ditch
		return nil, newHTTPError(res)
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

	out := rawResp.Post[0].toPost(d)
	return &out, nil
}
