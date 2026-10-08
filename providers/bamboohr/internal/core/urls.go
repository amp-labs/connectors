package core

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const APIVersion = "v1"

// ObjectURL builds a BambooHR REST URL from the provider base URL and object path from metadata.
// ProviderInfo.BaseURL is https://{workspace}.bamboohr.com/api; the version segment is applied here.
func ObjectURL(baseURL, objectPath string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL, APIVersion, objectPath)
}
