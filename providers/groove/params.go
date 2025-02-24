package groove

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
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
	username string, password string,
	opts ...common.HeaderAuthClientOption,
) Option {
	return func(p *parameters) {
		p.WithBasicClient(ctx, client, username, password, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(p *parameters) {
		p.WithAuthenticatedClient(client)
	}
}
