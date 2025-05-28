package pardot

import (
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
	return &Adapter{
		Client:  client,
		BaseURL: info.BaseURL,
	}, nil
}
