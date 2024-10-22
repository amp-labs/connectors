package customerapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

// DefaultPageSize is number of elements per page.
const DefaultPageSize = 50

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	apiKey string, opts ...common.HeaderAuthClientOption,
) Option {
	return func(params *parameters) {
		params.WithApiKeyHeaderClient(ctx, client, providers.CustomerJourneysApp, apiKey, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}
