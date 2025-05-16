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
	client *common.JSONHTTPClient, info *providers.ProviderInfo, metadata map[string]string,
) (*Adapter, error) {
	modules := info.Modules
	if modules == nil {
		return nil, common.ErrImplementation
	}

	baseURL := (*modules)[providers.ModuleSalesforceAccountEngagement].BaseURL
	if baseURL == "" {
		return nil, common.ErrInvalidModuleDeclaration
	}

	subdomain := getSubdomain(metadata)

	// TODO replace with proper ModuleInfo resolver.
	baseURL = strings.Replace(baseURL, "<<SERVICE_DOMAIN>>", subdomain, 1)

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
