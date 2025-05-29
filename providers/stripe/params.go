package stripe

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
)

const (
	// DefaultPageSize is number of elements per page.
	DefaultPageSize = 100
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parametersInternal)

// parametersInternal surface options by delegation.
type parametersInternal struct {
	paramsbuilder.Client
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		AuthenticatedClient: oldParams.Client.Caller.Client,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(
	ctx context.Context, client *http.Client,
	apiKey string,
	opts ...common.HeaderAuthClientOption,
) Option {
	return func(params *parametersInternal) {
		params.WithApiKeyHeaderClient(ctx, client, providers.Stripe, apiKey, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parametersInternal) {
		params.WithAuthenticatedClient(client)
	}
}
