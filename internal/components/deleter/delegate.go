package deleter

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

type DelegateDeleter struct {
	execute func(context.Context, common.DeleteParams) (*common.DeleteResult, error)
}

// NewDelegateDeleter creates a new DelegateDeleter with the provided delete function.
func NewDelegateDeleter(
	deleteFunc func(context.Context, common.DeleteParams) (*common.DeleteResult, error),
) *DelegateDeleter {
	return &DelegateDeleter{
		execute: deleteFunc,
	}
}

// Delete performs the delete operation.
func (d *DelegateDeleter) Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return d.execute(ctx, params)
}
