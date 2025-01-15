package components

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type DeleteStrategy interface {
	DeleteObject(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error)
}

func (c *ConnectorComponent) Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error) {
	if c.DeleteStrategy == nil {
		return nil, fmt.Errorf("%w: delete", common.ErrNotImplemented)
	}

	if len(params.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	support, err := c.GetSupport(c.module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// TODO: This is a placeholder support level for delete operation
	if !support.BulkWrite.Delete {
		return nil, common.ErrOperationNotSupportedForObject
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return c.DeleteStrategy.DeleteObject(ctx, params)
}
