package booru

import (
	"fmt"
	"net/url"
	"strings"
)

// queryify converts a key-value map into a URL query.
// It does not contain the preceeding "?".
// If kv is nil, an empty string is returned instead of panicking.
func queryify(kv map[string]string) string {
	if kv == nil {
		return ""
	}

	objs := make([]string, 0, len(kv))
	for k, v := range kv {
		if v != "" {
			objs = append(objs, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		} else {
			objs = append(objs, url.QueryEscape(k))
		}
	}

	return strings.Join(objs, "&")
}
