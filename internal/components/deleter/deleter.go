package deleter

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
)

type HTTPDeleter struct {
	operation *operations.DeleteOperation
	registry  *components.EndpointRegistry
	module    common.ModuleID
}

func NewHTTPDeleter(
	client common.AuthenticatedHTTPClient,
	registry *components.EndpointRegistry,
	module common.ModuleID,
	list operations.DeleteHandlers,
) *HTTPDeleter {
	return &HTTPDeleter{
		operation: operations.NewHTTPOperation(client, list),
		registry:  registry,
		module:    module,
	}
}

// Delete performs the delete operation.
func (d *HTTPDeleter) Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if d.operation == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrNotImplemented, "deleter is not implemented")
	}

	if d.registry == nil {
		return nil, components.ErrSupportNotConfigured
	}

	support, err := d.registry.GetSupport(d.module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// TODO: Add support.Delete
	if !support.BulkWrite.Delete {
		return nil, fmt.Errorf("%w: %s does not support delete", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return d.operation.ExecuteRequest(ctx, params)
}
