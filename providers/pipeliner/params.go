package pipeliner

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
)

const (
	// DefaultPageSize is number of elements per page.
	DefaultPageSize = 100
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parametersInternal)

// parameters surface options by delegation.
type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	user, pass string, opts ...common.HeaderAuthClientOption,
) Option {
	return func(params *parametersInternal) {
		params.WithBasicClient(ctx, client, user, pass, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parametersInternal) {
		params.WithAuthenticatedClient(client)
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *parametersInternal) {
		params.WithWorkspace(workspaceRef)
	}
}
