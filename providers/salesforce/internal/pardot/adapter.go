package pardot

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	Client  *common.JSONHTTPClient
	BaseURL string
}

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo, metadata map[string]string,
) (*Adapter, error) {
	subdomain := getSubdomain(metadata)

	baseURL := strings.Replace(info.BaseURL, "<<SERVICE_DOMAIN>>", subdomain, 1)

	return &Adapter{
		Client:  client,
		BaseURL: baseURL,
	}, nil
}

func getSubdomain(metadata map[string]string) string {
	isDemoValue, ok := metadata["isDemo"]
	if !ok {
		return ""
	}

	if strings.ToLower(isDemoValue) == "true" {
		return ".demo"
	}

	return ""
}
