package servicedeskplus

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "api/v3"

type Adapter struct {
	Client  *common.JSONHTTPClient
	BaseURL string
}

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo, metadata map[string]string,
) (*Adapter, error) {
	return &Adapter{
		Client:  client,
		BaseURL: info.BaseURL,
	}, nil
}

func (a *Adapter) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, apiVersion, objectName)
}
