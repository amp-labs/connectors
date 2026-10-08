package connectwise

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
)

var _ connectors.BatchRecordReaderConnector = (*Connector)(nil)

// GetRecordsByIds scoped reading of records given their ids.
// nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) ([]common.ReadResultRow, error) {
	if len(recordIds) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Ensure identifiers are non-repeating.
	ids := datautils.NewSetFromList(recordIds).List()

	batchResult, err := batch.Read[map[string]any](ctx, c.batchAdapter, objectName, ids)
	if err != nil {
		return nil, err
	}

	marshaler := readhelper.MakeMarshaledDataFuncWithId(
		recordTransformer(objectName),
		readhelper.IdFieldQuery{Field: "id"},
	)
	uniqueFields := datautils.NewSetFromList(fields).List()

	list, err := jsonquery.Convertor.NodesFromArray(batchResult)
	if err != nil {
		return nil, err
	}

	return marshaler(list, uniqueFields)
}
