package smartlead

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

// parameters surface options by delegation.
type parameters struct {
	paramsbuilder.Client
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	apiKey string,
	opts ...common.QueryParamAuthClientOption,
) Option {
	return func(params *parameters) {
		params.WithApiKeyQueryParamClient(ctx, client, providers.Smartlead, apiKey, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}
