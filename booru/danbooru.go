package booru

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Danbooru implements the Danbooru API.
//
// API documentation: https://danbooru.donmai.us/wiki_pages/help:api
type Danbooru struct {
	// URL is the location of where the Danbooru API is.
	// This is a necessary field, or else all requests will fail as they have
	// nowhere to go.
	// URL must not change after first use.
	//
	// Example: "https://safebooru.donmai.us"
	URL *url.URL

	// HttpClient is the HTTP client object that is used to talk to the
	// Danbooru API.
	HttpClient *http.Client
}

// danbooruPost holds some of the information returned by the Danbooru API.
// This isn't supposed to be used outside of this package; it is simply here to
// ease unmarshaling of responses.
// Always convert to the standard Post struct instead.
type danbooruPost struct {
	Id int

	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`

	Score int

	Source string

	Ext         string `json:"file_ext"`
	Size        int    `json:"file_size"`
	OriginalUrl string `json:"file_url"`
	ThumbUrl    string `json:"large_file_url"`
	Width       int    `json:"image_width"`
	Height      int    `json:"image_height"`

	Tags string `json:"tag_string"`

	Rating string
}

// toPost converts the internal representation to an actual Post used by the
// outer world.
func (dp danbooruPost) toPost(d *Danbooru) Post {
	// Luckily for us, there's a rather direct conversion.

	var r Rating
	switch dp.Rating {
	default:
		fallthrough
	case "g":
		r = General
	case "q":
		r = Questionable
	case "s":
		r = Sensitive
	case "e":
		r = Explicit
	}

	return Post{
		Id:      dp.Id,
		Score:   dp.Score,
		Source:  dp.Source,
		Created: dp.Created,
		Updated: dp.Updated,
		Tags:    strings.Split(dp.Tags, " "),
		Rating:  r,
		Original: Image{
			Href:   dp.OriginalUrl,
			MIME:   mime.TypeByExtension("." + dp.Ext), // mime asks we include the dot
			Size:   dp.Size,
			Width:  dp.Width,
			Height: dp.Height,
		},
		Thumbnail: Image{
			Href: dp.ThumbUrl,
			MIME: "image/jpeg", // assumption
			Size: 0,            // we are never told
		},
		Origin: d,
	}
}

// HTTP returns the HttpClient that this booru uses.
func (d *Danbooru) HTTP() *http.Client {
	return d.HttpClient
}

func (d *Danbooru) Page(ctx context.Context, q Query, page int) ([]Post, int, error) {
	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = path.Join(u.Path, "posts.json")
	uq := u.Query()
	uq.Set("page", fmt.Sprint(page))
	uq.Set("tags", strings.Join(q.Tags, " "))
	u.RawQuery = uq.Encode()

	// Create a request object
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, -1, err
	}

	// Do the needful
	res, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Something bad happened, ditch
		return nil, -1, newHTTPError(res)
	}

	// Parse the results
	var rawList []danbooruPost
	if err := json.NewDecoder(res.Body).Decode(&rawList); err != nil {
		return nil, -1, err
	}

	// Convert
	out := make([]Post, len(rawList))
	for i, v := range rawList {
		out[i] = v.toPost(d)
	}

	return out, -1, nil
}

func (d *Danbooru) Post(ctx context.Context, id int) (*Post, error) {
	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = path.Join(u.Path, fmt.Sprintf("posts/%d.json", id))

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
		return nil, newHTTPError(res)
	}

	// Parse the result
	var rawPost danbooruPost
	if err := json.NewDecoder(res.Body).Decode(&rawPost); err != nil {
		return nil, err
	}

	out := rawPost.toPost(d)
	return &out, nil
}
