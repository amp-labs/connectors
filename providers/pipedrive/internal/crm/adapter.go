package crm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	enum               = "enum"
	set                = "set"
	notes              = "notes"
	activities         = "activities"
	deals              = "deals"
	products           = "products"
	organizations      = "organizations"
	persons            = "persons"
	pipelines          = "pipelines"
	stages             = "stages"
	metadataAPIVersion = "v1"
)

type Adapter struct {
	Client  *common.JSONHTTPClient
	BaseURL string
}

func NewAdapter(
	client *common.JSONHTTPClient, baseURL string,
) *Adapter {
	return &Adapter{
		Client:  client,
		BaseURL: baseURL,
	}
}

func (a *Adapter) getAPIURL(apiVersion, object string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, apiVersion, object)
}

func (a *Adapter) constructMetadataURL(obj string) (*urlbuilder.URL, error) {
	if metadataDiscoveryEndpoints.Has(obj) {
		obj = metadataDiscoveryEndpoints[obj]
	}

	return a.getAPIURL(metadataAPIVersion, obj)
}
