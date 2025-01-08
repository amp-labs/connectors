package components

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type WriteStrategy interface {
	WriteObject(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
}

func (c *ConnectorComponent) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if c.WriteStrategy == nil {
		return nil, fmt.Errorf("%w: write", common.ErrNotImplemented)
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

	if !support.Write {
		return nil, common.ErrOperationNotSupportedForObject
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return c.WriteStrategy.WriteObject(ctx, params)
}
