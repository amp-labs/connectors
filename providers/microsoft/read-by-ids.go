package microsoft

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
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

	batchParams, requestIdentifiers, err := c.paramsForBatchRead(objectName, recordIds)
	if err != nil {
		return nil, err
	}

	batchResponse := batch.Execute[map[string]any](ctx, c.batchStrategy, batchParams)
	if err = batchResponse.JoinedErr(); err != nil {
		return nil, err
	}

	marshaler := readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id"))
	uniqueFields := datautils.NewSetFromList(fields).List()

	return marshaler(batchResponse.GetInOrder(requestIdentifiers), uniqueFields)
}

func (c *Connector) paramsForBatchRead(
	objectName string, identifiers []string,
) (*batch.Params, []batch.RequestID, error) {
	batchParams := &batch.Params{}

	requestIdentifiers := make([]batch.RequestID, len(identifiers))
	for index, identifier := range identifiers {
		url, err := c.getURL(objectName)
		if err != nil {
			return nil, nil, err
		}

		url.AddPath(identifier)
		requestIdentifier := batch.RequestID(fmt.Sprintf("%v_%v", objectName, identifier))
		requestIdentifiers[index] = requestIdentifier
		batchParams.WithRequest(requestIdentifier, http.MethodGet, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, requestIdentifiers, nil
}
