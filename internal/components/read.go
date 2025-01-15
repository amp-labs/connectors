package components

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type ReadStrategy interface {
	ReadObject(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
}

func (c *ConnectorComponent) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if c.ReadStrategy == nil {
		return nil, fmt.Errorf("%w: read", common.ErrNotImplemented)
	}

	if len(params.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	support, err := c.GetSupport(c.module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !support.Read {
		return nil, common.ErrOperationNotSupportedForObject
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return c.ReadStrategy.ReadObject(ctx, params)
}
