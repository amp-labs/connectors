package freshdesk

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
	paramsbuilder.Workspace
}

func newParams(opts []Option) (*common.Parameters, error) {
	oldParams, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	return &common.Parameters{
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
	}, nil
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

// WithWorkspace sets the freshdesk API instance to use for the connector. It's required.
func WithWorkspace(workspace string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspace)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(p *parameters) {
		p.WithAuthenticatedClient(client)
	}
}
