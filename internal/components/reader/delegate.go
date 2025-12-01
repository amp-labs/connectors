package reader

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// DelegateReader implements Reader by delegating to a provided function.
// Use this when the connector needs full control over the read implementation
// (e.g., SQL-based connectors like Snowflake).
type DelegateReader struct {
	execute func(context.Context, common.ReadParams) (*common.ReadResult, error)
}

// NewDelegateReader creates a new DelegateReader with the provided read function.
func NewDelegateReader(
	execute func(context.Context, common.ReadParams) (*common.ReadResult, error),
) *DelegateReader {
	return &DelegateReader{execute: execute}
}

func (r *DelegateReader) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	return r.execute(ctx, params)
}
