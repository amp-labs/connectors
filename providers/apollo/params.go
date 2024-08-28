package apollo

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

var headerName = "X-Api-Key" //nolint:gochecknoglobals

type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(
	ctx context.Context, client *http.Client,
	apiKey string,
	opts ...common.HeaderAuthClientOption,
) Option {
	options := []common.HeaderAuthClientOption{
		common.WithHeaderClient(client),
	}

	apiKeyClient, err := common.NewApiKeyHeaderAuthHTTPClient(ctx,
		headerName, apiKey,
		append(options, opts...)...,
	)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	return WithAuthenticatedClient(apiKeyClient)
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}
