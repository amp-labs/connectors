package writer

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
)

var _ components.Writer = &HTTPWriter{}

// HTTPWriter implements Writer using HTTP.
type HTTPWriter struct {
	operation *operations.WriteOperation
	registry  *components.EndpointRegistry
	module    common.ModuleID
}

func NewHTTPWriter(
	client common.AuthenticatedHTTPClient,
	registry *components.EndpointRegistry,
	module common.ModuleID,
	list operations.WriteHandlers,
) *HTTPWriter {
	return &HTTPWriter{
		operation: operations.NewHTTPOperation(client, list),
		registry:  registry,
		module:    module,
	}
}

func (w *HTTPWriter) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if w.operation == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrNotImplemented, "writer is not implemented")
	}

	// If there's no support, we can't validate the operation.
	if w.registry == nil {
		return nil, components.ErrSupportNotConfigured
	}

	support, err := w.registry.GetSupport(w.module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !support.Write {
		return nil, fmt.Errorf("%w: %s does not support write", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return w.operation.ExecuteRequest(ctx, params)
}
