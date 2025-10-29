package batch

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewAdapter(httpClient *common.HTTPClient, moduleInfo *providers.ModuleInfo) *Adapter {
	return &Adapter{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		moduleInfo: moduleInfo,
	}
}
