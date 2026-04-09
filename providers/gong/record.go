package gong

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds fetches full records from Gong for a specific set of IDs.
// For calls, uses POST /v2/calls/extensive with filter.callIds.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	if objectName != objectNameCalls {
		return nil, fmt.Errorf("%w: gong only supports %q, got %q",
			common.ErrGetRecordNotSupportedForObject, objectNameCalls, objectName)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	url, err := c.getReadURL(objectName)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"filter": map[string]any{
			"callIds": ids,
		},
		"contentSelector": callContentSelector(),
	}

	response, err := c.Client.Post(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectName)

	readResult, err := common.ParseResult(response,
		getRecords(responseFieldName),
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(extractMetaDataFields),
		datautils.NewSetFromList(fields),
	)
	if err != nil {
		return nil, err
	}

	return readResult.Data, nil
}
