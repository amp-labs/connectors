package apollo

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
	return func(params *parameters) {
		params.WithApiKeyHeaderClient(ctx, client, providers.Apollo, apiKey, opts...)
	}
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
