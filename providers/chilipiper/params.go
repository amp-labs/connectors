package chilipiper

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
)

type Option = func(params *parametersInternal)

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
	return func(p *parametersInternal) {
		p.WithApiKeyHeaderClient(ctx, client, providers.ChiliPiper, apiKey, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(p *parametersInternal) {
		p.WithAuthenticatedClient(client)
	}
}
