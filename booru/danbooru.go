package booru

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Danbooru implements the Danbooru API.
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
	Updated time.Time `json:"created_at"`

	Score int

	Source string

	Ext         string `json:"file_ext"`
	Size        int    `json:"file_size"`
	OriginalUrl string `json:"file_url"`
	ThumbUrl    string `json:"preview_file_url"`

	Tags string `json:"tag_string"`
}

var danbooruErrors = map[int]error{
	204: errors.New("danbooru: no content (204)"),
	403: errors.New("danbooru: forbidden (403)"),
	404: errors.New("danbooru: not found (404)"),
	420: errors.New("danbooru: record could not be saved (420)"),
	421: errors.New("danbooru: user throttled (421)"),
	422: errors.New("danbooru: locked (422)"),
	423: errors.New("danbooru: already exists (423)"),
	424: errors.New("danbooru: invalid parameters (424)"),
	500: errors.New("danbooru: internal server error (500)"),
	503: errors.New("danbooru: unavailable (503)"),
}

// toPost converts the internal representation to an actual Post used by the
// outer world.
func (dp danbooruPost) toPost() Post {
	// Luckily for us, there's a rather direct conversion.
	return Post{
		Id:      dp.Id,
		Score:   dp.Score,
		Source:  dp.Source,
		Created: dp.Created,
		Updated: dp.Updated,
		Tags:    strings.Split(dp.Tags, " "),
		Images: []Image{
			{
				Href:      dp.OriginalUrl,
				MIME:      mime.TypeByExtension("." + dp.Ext), // mime asks we include the dot
				Size:      dp.Size,
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
}

func (d *Danbooru) Page(ctx context.Context, q Query, page int) ([]Post, error) {
	urlq := queryify(map[string]string{
		"page": fmt.Sprint(page),
		"tags": strings.Join(q.Tags, " "),
	})

	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = filepath.Join(u.Path, "posts.json")

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
		if err, ok := danbooruErrors[res.StatusCode]; ok {
			return nil, err
		}

		return nil, fmt.Errorf("danbooru: unknown error (%d)", res.StatusCode)
	}

	// Parse the results
	var rawList []danbooruPost

	if err := json.NewDecoder(res.Body).Decode(&rawList); err != nil {
		return nil, err
	}

	// Convert
	out := make([]Post, len(rawList))
	for i, v := range rawList {
		out[i] = v.toPost()
	}

	return out, nil
}

func (d *Danbooru) Post(ctx context.Context, id int) (*Post, error) {
	// Copy our URL object so we can set the query
	u := *d.URL

	u.Path = filepath.Join(u.Path, fmt.Sprintf("posts/%d.json", id))

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
		if err, ok := danbooruErrors[res.StatusCode]; ok {
			return nil, err
		}

		return nil, fmt.Errorf("danbooru: unknown error (%d)", res.StatusCode)
	}

	// Parse the results
	var rawPost danbooruPost

	if err := json.NewDecoder(res.Body).Decode(&rawPost); err != nil {
		return nil, err
	}

	out := rawPost.toPost()
	return &out, nil
}
