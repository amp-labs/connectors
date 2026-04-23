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
// Supported objects:
//   - calls: POST /v2/calls/extensive with filter.callIds
//   - users: POST /v2/users/extensive with filter.userIds
//
// https://gong.app.gong.io/settings/api/documentation#overview
func (c *Connector) GetRecordsByIds( //nolint:revive,funlen
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var (
		filterKey       string
		contentSelector map[string]any
		transformer     common.RecordTransformer
		extraPath       string
	)

	switch objectName {
	case objectNameCalls:
		// getReadURL already appends /extensive for calls.
		filterKey = "callIds"
		contentSelector = callContentSelector()
		transformer = extractMetaDataFields
	case objectNameUsers:
		filterKey = "userIds"
		extraPath = "extensive" // this endpoint allows filtering by userIds
	default:
		return nil, fmt.Errorf("%w: gong supports %q and %q, got %q",
			common.ErrGetRecordNotSupportedForObject,
			objectNameCalls, objectNameUsers, objectName)
	}

	url, err := c.getReadURL(objectName)
	if err != nil {
		return nil, err
	}

	if extraPath != "" {
		url.AddPath(extraPath)
	}

	body := map[string]any{
		"filter": map[string]any{filterKey: ids},
	}

	if contentSelector != nil {
		body["contentSelector"] = contentSelector
	}

	response, err := c.Client.Post(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectName)

	readResult, err := common.ParseResult(response,
		getRecords(responseFieldName),
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(transformer),
		datautils.NewSetFromList(fields),
	)
	if err != nil {
		return nil, err
	}

	return readResult.Data, nil
}
