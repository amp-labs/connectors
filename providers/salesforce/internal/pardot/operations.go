package pardot

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return Schemas.Select(objectNames)
}

func (a *Adapter) GetRecordCount(
	ctx context.Context, params *common.RecordCountParams,
) (*common.RecordCountResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}
