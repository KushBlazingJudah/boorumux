package booru

import (
	"context"
	"net/http"
	"time"
)

// API is an interface that neatly abstracts the hard parts of parsing and
// mangling results from the actual booru's API into something easy to use and
// comprehend.
type API interface {
	// Page returns a page from this booru's API.
	// The number of posts returned is arbitrary; depending on the filters
	// used, there may be less than an expected amount.
	Page(ctx context.Context, q Query, page int) ([]Post, error)

	// Post returns a specific post referenced by its numeric ID.
	// The post returned will be nil when error is not.
	Post(ctx context.Context, id int) (*Post, error)

	// HTTP returns the HTTP client that this booru uses.
	HTTP() *http.Client
}

// Query is a list of options passed to a booru API that are used as a query to
// fetch results.
type Query struct {
	// Tags are used to further target desired topics or even characters.
	// They can be also used in the opposite way to exclude topics by being
	// preceeded with a "-".
	Tags []string
}

// Image contains enough data to uniquely identify this image and to download it.
type Image struct {
	// Href is the link to this specific image.
	// It should be able to be downloaded from this specific link at any given
	// time.
	Href string

	// MIME is the MIME-type of this specific image.
	MIME string

	// Size is the file size to this image.
	// This value may be 0, in which case the file size is unknown.
	Size int

	// Thumbnail is an indicator for if this image is supposed to be used as an
	// thumbnail, and is not the full, high resolution image.
	Thumbnail bool
}

// Post contains data related to a specific post on any given booru.
type Post struct {
	// Id is a numeric id that corresponds to this exact post.
	Id int

	// Score is the numeric score of this specific post.
	Score int

	// Source is the source URL for this post.
	// This value may be empty, in which case the source is unknown.
	Source string

	// Created is the time that this post was created.
	Created time.Time

	// Updated is the time that this post was updated.
	Updated time.Time

	// Tags is a list of tags associated with this specific post.
	Tags []string

	// Images holds the various images for this post, such as the original
	// image uploaded to the booru or the thumbnail.
	Images []Image
}
