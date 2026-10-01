package confluence

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}
