package apollo

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var headerName = "X-Api-Key" //nolint:gochecknoglobals

type Option = func(params *parameters)

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

func usesSearching(objectName string) bool {
	return in(objectName, postSearchObjects, getSearchObjects)
}

func in(a string, b ...[]ObjectType) bool {
	o := ObjectType(a)

	for _, sl := range b {
		for _, v := range sl {
			if v == o {
				return true
			}
		}
	}

	return false
}
