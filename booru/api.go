package booru

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("booru: not found")
)

type Rating int

const (
	General Rating = iota
	Questionable
	Sensitive
	Explicit
)

// API is an interface that neatly abstracts the hard parts of parsing and
// mangling results from the actual booru's API into something easy to use and
// comprehend.
type API interface {
	// Page returns a page from this booru's API according to a query.
	//
	// The number of posts returned is arbitrary; depending on the filters
	// used, there may be less than an expected amount.
	//
	// The integer returned by this function should be the number of remaining
	// pages, however in the case where it is unknown -1 will be used.
	Page(ctx context.Context, q Query, page int) ([]Post, int, error)

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

	// Width and Height represent the dimensions of the image, which may be 0
	// if unknown.
	Width, Height int
}

// Post contains data related to a specific post on any given booru.
type Post struct {
	// Id is a numeric id that corresponds to this exact post.
	Id int

	// Score is the numeric score of this specific post.
	Score int

	// Source is the source location for this post.
	// This value may be empty, in which case the source is unknown.
	// It is usually a URL; if it is, it should be treated as such.
	Source string

	// Created is the time that this post was created.
	Created time.Time

	// Updated is the time that this post was updated.
	Updated time.Time

	// Tags is a list of tags associated with this specific post.
	Tags []string

	// Original is the original or highest quality picture for this post.
	Original Image

	// Thumbnail is a thumbnail for this post.
	// This is normally the "large" but not "original" size of an image.
	Thumbnail Image

	Rating Rating
	Origin API
}

// HTTPError represents a generic HTTP failure status code message using the
// error interface.
type HTTPError struct {
	// URL is the URL that was requested that returned this error.
	URL string

	// Code is the status code returned by the server.
	Code int
}

// IsVideo determines if this image is a video by looking at its MIME type.
func (i Image) IsVideo() bool {
	return strings.HasPrefix(i.MIME, "video/")
}

func (h HTTPError) Error() string {
	var msg string
	switch h.Code {
	case 204:
		msg = "no content"
	case 403:
		msg = "forbidden"
	case 404:
		msg = "not found"
	case 420:
		msg = "record could not be saved"
	case 421:
		msg = "user throttled"
	case 422:
		msg = "locked"
	case 423:
		msg = "already exists"
	case 424:
		msg = "invalid parameters"
	case 500:
		msg = "internal server error"
	case 503:
		msg = "unavailable"
	default:
		msg = "unknown"
	}

	return fmt.Sprintf("booru: %s returned status %d: %s", h.URL, h.Code, msg)
}

func newHTTPError(req *http.Response) HTTPError {
	return HTTPError{
		URL:  req.Request.URL.String(),
		Code: req.StatusCode,
	}
}
