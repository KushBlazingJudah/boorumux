package filter

import (
	"strings"

	"github.com/KushBlazingJudah/boorumux/booru"
)

func matchRating(c string, p *booru.Post) bool {
	for _, r := range strings.Split(c, "|") {
		var rr booru.Rating
		switch r {
		default:
			fallthrough
		case "general", "safe", "g", "sfw":
			rr = booru.General
		case "questionable", "q":
			rr = booru.Questionable
		case "sensitive", "s":
			rr = booru.Sensitive
		case "explicit", "e":
			rr = booru.Explicit
		}

		if p.Rating == rr {
			return true
		}
	}

	return false
}

func matchTag(ts string, p *booru.Post) bool {
	for _, t := range strings.Split(ts, "|") {
		for _, v := range p.Tags {
			if v == string(t) {
				return true
			}
		}
	}

	return false
}

func (f Filter) Match(p *booru.Post) bool {
	for _, v := range f {
		neg := strings.HasPrefix(v, "-")
		if neg {
			v = strings.TrimPrefix(v, "-")
		}

		var res bool

		if strings.HasPrefix(v, "rating:") {
			r := strings.TrimPrefix(v, "rating:")
			res = matchRating(r, p)
			goto check
		}

		// Just match against the tags
		res = matchTag(v, p)

	check:
		if neg {
			res = !res
		}

		if !res {
			return false
		}
	}

	return true
}
