package breezy

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersion       = "v3"
	companyIDPlaceholder = "{company_id}"
)

func buildVersionedPathURL(baseURL string, path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL, restAPIVersion, strings.TrimSpace(path))
}

func resolveObjectPath(path string, companyID string) string {
	return strings.ReplaceAll(path, companyIDPlaceholder, companyID)
}
