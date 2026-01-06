package hubspot

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

const (
	// DefaultPageSize is the default page size for paginated requests.
	// HubSpot's read endpoints support max 100 records per page.
	// Note: Search endpoints support up to 200, but we use a shared default.
	DefaultPageSize    = "100"
	DefaultPageSizeInt = 100
)

// Option is a function which mutates the hubspot connector configuration.
type Option = func(params *parameters)

// parameters is the internal configuration for the hubspot connector.
type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Module
}

func newParams(opts []Option) (*common.ConnectorParams, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(common.ModuleRoot),
	)
	if err != nil {
		return nil, err
	}

	return &common.ConnectorParams{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
	}, nil
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Module.ValidateParams(),
	)
}

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

// WithModule sets the hubspot API module to use for the connector. It's required.
func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, common.ModuleRoot)
	}
}
