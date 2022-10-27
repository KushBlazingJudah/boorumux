package boorumux

import (
	"fmt"
	"html/template"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// counter can count items according to how often they appear.
//
// counter is not thread safe.
type counter map[interface{}]int

var counterPool = sync.Pool{
	New: func() any {
		return counter{}
	},
}

var (
	schemaRegexp = regexp.MustCompile("^https?://")
)

func (c counter) count(a interface{}) {
	n, ok := c[a]
	if !ok {
		c[a] = 1
	} else {
		c[a] = n + 1
	}
}

func (c counter) reset() {
	for k := range c {
		delete(c, k)
	}
}

func humanSize(b int) string {
	var f float64
	var r string
	if b > 1024*1024 {
		f = float64(b) / (1024 * 1024)
		r = "MB"
	} else if b > 1024 {
		f = float64(b) / 1024
		r = "KB"
	} else {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.2f %s", f, r)
}

// mostCommon returns the most common items in a list.
func mostCommon(list []string) []string {
	c := counterPool.Get().(counter)

	// defers are LIFO; this resets it first and then checks it into the pool
	defer counterPool.Put(c)
	defer c.reset()

	for _, v := range list {
		c.count(v)
	}

	o := make([]string, 0, len(c))
	for k := range c {
		o = append(o, k.(string))
	}

	sort.Strings(o)

	sort.Slice(o, func(i, j int) bool {
		return c[o[i]] < c[o[j]]
	})

	for i, j := 0, len(o)-1; i < j; i, j = i+1, j-1 {
		o[i], o[j] = o[j], o[i]
	}

	return o
}

func buildPageBlock(base string, current int) template.HTML {
	// TODO: This function really sucks; list out a couple pages, in a form
	// similar to this: 1 2 ... 5 [6] 7 ... 12

	sb := strings.Builder{}
	sb.WriteString(`<div id="pages">`)
	if current != 0 {
		fmt.Fprintf(&sb, ` <a href="%s">0</a>`, base)
		if current-1 > 0 {
			fmt.Fprintf(&sb, ` ... <a href="%s&page=%d">%d</a> `, base, current-1, current-1)
		}
		fmt.Fprintf(&sb, ` <b>%d</b> `, current)
	}
	fmt.Fprintf(&sb, ` <a href="%s&page=%d">next</a>`, base, current+1)
	sb.WriteString(`</div>`)
	return template.HTML(sb.String())
}

func prettyUrl(u string) string {
	return schemaRegexp.ReplaceAllString(u, "")
}
