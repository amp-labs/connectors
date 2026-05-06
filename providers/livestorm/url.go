package livestorm

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const apiVersion = "v1"

func buildURLWithVersion(baseURL string, path string, segments ...string) (*urlbuilder.URL, error) {
	parts := []string{apiVersion}

	trimmedPath := strings.TrimSpace(path)
	if trimmedPath != "" {
		for _, p := range strings.Split(strings.Trim(trimmedPath, "/"), "/") {
			if p != "" {
				parts = append(parts, p)
			}
		}
	}

	parts = append(parts, segments...)

	return urlbuilder.New(baseURL, parts...)
}
