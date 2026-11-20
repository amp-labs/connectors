package closecrm

import "strings"

// addTrailingSlashIfNeeded ensures the URL path ends with a slash.
// The slash belongs to the path, so it is placed before the query string if there is one.
// Appending it to the whole URL would corrupt the last query parameter (e.g. `_skip=0/`).
func addTrailingSlashIfNeeded(urlString string) string {
	if len(urlString) == 0 {
		return urlString
	}

	path, query, hasQuery := strings.Cut(urlString, "?")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	if hasQuery {
		return path + "?" + query
	}

	return path
}
