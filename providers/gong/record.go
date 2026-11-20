package gong

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds fetches full records from Gong for a specific set of IDs.
// Supported objects:
//   - calls: POST /v2/calls/extensive with filter.callIds
//   - users: POST /v2/users/extensive with filter.userIds
//
// Every page of the response is read. Requested ids that Gong does not return are reported as
// not-found failures.
//
// https://gong.app.gong.io/settings/api/documentation#overview
func (c *Connector) GetRecordsByIds( //nolint:revive,funlen
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	_ []string,
) (*common.BatchReadResult, error) {
	if len(ids) == 0 {
		return common.NewBatchReadResult(nil), nil
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
		contentSelector = callContentSelector(datautils.NewSetFromList(fields))
		transformer = extractMetaDataFields
	case objectNameUsers:
		filterKey = "userIds"
		extraPath = "extensive" // this endpoint allows filtering by userIds
	default:
		return nil, fmt.Errorf("%w: %q", common.ErrGetRecordNotSupportedForObject, objectName)
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

	rows, err := c.postAllPages(ctx, url, body, objectName, transformer, datautils.NewSetFromList(fields))
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult(rows)
	result.ReportMissingIds(ids)

	return result, nil
}

// postAllPages sends body to url, then re-sends it with each page's records.cursor until a page has
// none, and returns the rows of every page.
func (c *Connector) postAllPages(
	ctx context.Context,
	url *urlbuilder.URL,
	body map[string]any,
	objectName string,
	transformer common.RecordTransformer,
	fields datautils.Set[string],
) ([]common.ReadResultRow, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectName)

	var rows []common.ReadResultRow

	for {
		response, err := c.Client.Post(ctx, url.String(), body)
		if err != nil {
			// Gong documents a 404 from calls/extensive as "No calls found for the specified period",
			// so none of the ids not yet returned exist. Read treats this 404 the same way.
			// https://help.gong.io/apidocs/retrieve-detailed-call-data-by-various-filters-v2callsextensive-2.md
			if errors.Is(err, common.ErrNotFound) {
				return rows, nil
			}

			return nil, err
		}

		readResult, err := common.ParseResult(response,
			getRecords(responseFieldName),
			getNextRecordsURL,
			common.MakeMarshaledDataFunc(transformer),
			fields,
		)
		if err != nil {
			return nil, err
		}

		rows = append(rows, readResult.Data...)

		if readResult.NextPage == "" {
			return rows, nil
		}

		body["cursor"] = readResult.NextPage.String()
	}
}
