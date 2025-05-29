package keap

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// DefaultPageSize is number of elements per page.
const DefaultPageSize = 50

// Option is a function which mutates the connector configuration.
type Option = func(params *parametersInternal)

type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Module
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts,
		WithModule(providers.ModuleKeapV1),
	)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Module.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parametersInternal) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parametersInternal) {
		params.WithAuthenticatedClient(client)
	}
}

func WithModule(module common.ModuleID) Option {
	return func(params *parametersInternal) {
		params.WithModule(module, SupportedModules, providers.ModuleKeapV1)
	}
}
