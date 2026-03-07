package crm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	enum             = "enum"
	set              = "set"
	apiVersion       = "api/v2"
	metadataPageSize = 500
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

func (a *Adapter) getAPIURL(object string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleinfo.BaseURL, apiVersion, object)
}

func (a *Adapter) constructMetadataURL(objectName string) (*urlbuilder.URL, error) {
	if metadataDiscoveryEndpoints.Has(objectName) {
		objectName = metadataDiscoveryEndpoints[objectName]
	}

	return a.getAPIURL(objectName)
}
