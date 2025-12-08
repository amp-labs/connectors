package legacy

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

const (
	apiVersion string = "v1" // nolint:gochecknoglobals
)

type Adapter struct {
	Client     *common.JSONHTTPClient
	moduleinfo *providers.ModuleInfo
}

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo,
) *Adapter {
	return &Adapter{
		Client:     client,
		moduleinfo: info,
	}
}

// getAPIURL constructs a specific object's resource URL in the format
// `{{baseURL}}/{{version}}/{{objectName}}`.
func (a *Adapter) getAPIURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleinfo.BaseURL, apiVersion, arg)
}

func (a *Adapter) constructMetadataURL(obj string) (*urlbuilder.URL, error) {
	if metadataDiscoveryEndpoints.Has(obj) {
		obj = metadataDiscoveryEndpoints[obj]
	}

	return a.getAPIURL(obj)
}

func (a *Adapter) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(providers.ModulePipedriveLegacy, objectName)
	if err != nil {
		return nil, err
	}

	return a.getAPIURL(path)
}
