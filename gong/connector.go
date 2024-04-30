package gong

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

var DefaultModule = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "api/data",
	Version: "v2",
}

type Connector struct {
	BaseURL   string
	Client    *common.JSONHTTPClient
	APIModule paramsbuilder.APIModule
}

func WithCatalogSubstitutions(substitutions map[string]string) Option {
	return func(params *gongParams) {
		params.substitutions = substitutions
	}
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer func() {
		if re := recover(); re != nil {
			tmp, ok := re.(error)
			if !ok {
				panic(re)
			}

			outErr = tmp
			conn = nil
		}
	}()

	params := &gongParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(providers.Gong, &map[string]string{
		"workspace": params.Workspace.Name,
	})
	if err != nil {
		return nil, err
	}

	return &Connector{
		Client:    params.client,
		BaseURL:   providerInfo.BaseURL,
		APIModule: params.APIModule,
	}, nil
}
