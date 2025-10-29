package batch

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// TODO implement batch write

func (a *Adapter) BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	return nil, common.ErrNotImplemented
}
