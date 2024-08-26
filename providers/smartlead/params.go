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
	// Error is set when any With<Method> fails, used for parameters validation.
	setupError error
}

func (p parameters) ValidateParams() error {
	if p.setupError != nil {
		return p.setupError
	}

	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	apiKey string,
	opts ...common.QueryParamAuthClientOption,
) Option {
	return func(params *parameters) {
		info, err := providers.ReadInfo(providers.Smartlead)
		if err != nil {
			params.setupError = err

			return
		}

		queryParam, err := info.GetApiKeyQueryParamName()
		if err != nil {
			params.setupError = err

			return
		}

		params.WithApiKeyQueryParamClient(ctx, client, queryParam, apiKey, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}
