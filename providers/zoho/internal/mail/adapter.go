package mail

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

// Adapter handles the Zoho Mail module.
//
// It currently provides metadata support only, sampling static-URL endpoints to
// infer object fields.
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

func (a *Adapter) getAPIURL(path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, path)
}
