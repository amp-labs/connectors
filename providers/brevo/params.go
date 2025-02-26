package brevo

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Module
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Module.ValidateParams(),
	)

}

func WithClient(ctx context.Context, client *http.Client,
	apiKey string, opts ...common.HeaderAuthClientOption,
) Option {

	return func(p *parameters) {
		p.WithApiKeyHeaderClient(ctx, client, providers.Brevo, apiKey, opts...)
	}

}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(p *parameters) {
		p.WithAuthenticatedClient(client)

	}
}
