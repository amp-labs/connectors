package pipedrive

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Module
}

func newParams(opts []Option) (*common.ConnectorParams, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	return &common.ConnectorParams{
		AuthenticatedClient: oldParams.Client.Caller.Client,
	}, nil
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, providers.ModulePipedriveLegacy)
	}
}

var supportedModules = common.Modules{ // nolint:gochecknoglobals
	providers.ModulePipedriveLegacy: common.Module{
		ID: providers.ModulePipedriveLegacy,
	},
	providers.ModulePipedriveCRM: common.Module{
		ID: providers.ModulePipedriveCRM,
	},
}
