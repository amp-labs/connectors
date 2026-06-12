package breezy

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersion       = "v3"
	companyIDPlaceholder = "{company_id}"
)

func buildVersionedPathURL(baseURL, objectPath string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL, restAPIVersion, strings.TrimSpace(objectPath))
}

func resolveObjectPath(objectPath, companyID string) string {
	if strings.Contains(objectPath, companyIDPlaceholder) {
		return strings.ReplaceAll(objectPath, companyIDPlaceholder, companyID)
	}

	return objectPath
}
