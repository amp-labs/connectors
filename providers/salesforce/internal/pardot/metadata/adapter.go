package metadata

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
)

type Strategy struct {
	// Basic connector
	*components.Connector

	xmlClient *common.XMLHTTPClient
}

func NewStrategy(base *components.Connector) (*Strategy, error) {
	return &Strategy{
		Connector: base,
		xmlClient: &common.XMLHTTPClient{
			HTTPClient: base.HTTPClient(),
		},
	}, nil
}

func (a *Strategy) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

func (a *Strategy) getSoapURL() (*urlbuilder.URL, error) {
	return urlbuilder.New("https://myself-9a-dev-ed.develop.soap.marketingcloudapis.com/Service.asmx")
}
