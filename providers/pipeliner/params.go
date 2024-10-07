package pipeliner

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

const (
	// DefaultPageSize is number of elements per page.
	DefaultPageSize = 100
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	user, pass string, opts ...common.HeaderAuthClientOption,
) Option {
	return func(params *parameters) {
		params.WithBasicClient(ctx, client, user, pass, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspaceRef)
	}
}
