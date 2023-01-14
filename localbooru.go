package boorumux

import (
	"net/http"
	"context"
	"mime/multipart"
	"bytes"
	"encoding/json"
	"strings"
	"mime"
	"time"
	"io"
	"fmt"
	"path"

	"github.com/KushBlazingJudah/boorumux/booru"
)

type lbPostinfo struct {
	Score  int      `json:"score"`
	Source string   `json:"source,omitempty"`
	Rating string   `json:"rating"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Booru string `json:"booru,omitempty"`
	BooruID string `json:"booru_id,omitempty"`
	Hash   string `json:"md5"`
	Ext    string `json:"file_ext"`
	Width  int    `json:"image_width"`
	Height int    `json:"image_height"`
	TagString string `json:"tag_string"`
}

func makePostinfo(bp *booru.Post) []byte {
	exts, _ := mime.ExtensionsByType(bp.Original.MIME)
	ext := exts[0][1:]

	pi := lbPostinfo{
		Score:   bp.Score,
		Source:  bp.Source,
		Created: bp.Created,
		Updated: bp.Updated,
		Hash:    bp.Hash,
		Ext: ext,
		Width: bp.Original.Width,
		Height: bp.Original.Height,
		TagString:    strings.Join(bp.Tags, " "),
	}

	// TODO: Booru, BooruID

	switch bp.Rating {
	default:
		fallthrough
	case booru.General:
		pi.Rating = "general"
	case booru.Questionable:
		pi.Rating = "questionable"
	case booru.Sensitive:
		pi.Rating = "sensitive"
	case booru.Explicit:
		pi.Rating = "explicit"
	}

	b, _ := json.Marshal(pi)
	return b
}

func (s *Server) save(ctx context.Context, post *booru.Post) error {
	buf := &bytes.Buffer{}
	mf := multipart.NewWriter(buf)

	// post info
	pi, err := mf.CreateFormField("info")
	if err != nil {
		return err
	}

	_, err = pi.Write(makePostinfo(post))
	if err != nil {
		return err
	}

	// the file itself
	pf, err := mf.CreateFormFile("file", path.Base(post.Original.Href))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", post.Original.Href, nil)
	if err != nil {
		return err
	}

	res, err := post.Origin.HTTP().Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if _, err := io.Copy(pf, res.Body); err != nil {
		return err
	}

	// finish
	mf.Close()

	// upload!
	resp, err := http.Post(fmt.Sprintf("%s/post", s.Localbooru), mf.FormDataContentType(), buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non 200 status code: %d", resp.StatusCode)
	}

	return nil
}
