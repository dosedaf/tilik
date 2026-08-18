package category

import (
	"strings"
	"net/url"
)

func extractKode(url *url.URL, prefix string) string {
	path := url.Path
	idx := strings.Index(path, prefix)
	if idx == -1 {
		return ""
	}

	remaining := path[idx+len(prefix):]

	return strings.Split(remaining, "/")[0]
}
