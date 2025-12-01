package writer

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// DelegateWriter implements Writer by delegating to a provided function.
// Use this when the connector needs full control over the write implementation
// (e.g., SQL-based connectors like Snowflake).
type DelegateWriter struct {
	execute func(context.Context, common.WriteParams) (*common.WriteResult, error)
}

// NewDelegateWriter creates a new DelegateWriter with the provided write function.
func NewDelegateWriter(
	execute func(context.Context, common.WriteParams) (*common.WriteResult, error),
) *DelegateWriter {
	return &DelegateWriter{
		execute: execute,
	}
}

func (w *DelegateWriter) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return w.execute(ctx, params)
}
