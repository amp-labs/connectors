package pardot

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	Client         *common.JSONHTTPClient
	BaseURL        string
	BusinessUnitID string
}

var ErrMissingBusinessUnitID = errors.New("missing metadata variable: business unit id")

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
	baseURL = strings.Replace(baseURL, "{{.subdomain}}", subdomain, 1)

	businessUnitID, ok := metadata["businessUnitId"]
	if !ok || businessUnitID == "" {
		return nil, ErrMissingBusinessUnitID
	}

	return &Adapter{
		Client:         client,
		BaseURL:        baseURL,
		BusinessUnitID: businessUnitID,
	}, nil
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, "api/v5/objects", objectName)
}

func (a *Adapter) businessUnitHeader() common.Header {
	return common.Header{
		Key:   "Pardot-Business-Unit-Id",
		Value: a.BusinessUnitID,
	}
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
