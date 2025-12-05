package reader

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
)

// HTTPReader implements a reading operation using HTTP, and uses the EndpointSupport
// to validate the operation.
type HTTPReader struct {
	operation *operations.ReadOperation
	registry  *components.EndpointRegistry
	module    common.ModuleID
}

func NewHTTPReader(
	client common.AuthenticatedHTTPClient,
	registry *components.EndpointRegistry,
	module common.ModuleID,
	list operations.ReadHandlers,
) *HTTPReader {
	return &HTTPReader{
		operation: operations.NewHTTPOperation(client, list),
		registry:  registry,
		module:    module,
	}
}

func (r *HTTPReader) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if r.operation == nil {
		return nil, fmt.Errorf("%w: reader is not implemented", common.ErrNotImplemented)
	}

	// If there's no support, we can't validate the operation.
	if r.registry == nil {
		return nil, components.ErrSupportNotConfigured
	}

	support, err := r.registry.GetSupport(r.module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !support.Read {
		return nil, fmt.Errorf("%w: %s does not support read", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return r.operation.ExecuteRequest(ctx, params)
}
