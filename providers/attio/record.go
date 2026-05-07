package attio

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds fetches records by their IDs for a given object type.
// Ref: https://docs.attio.com/rest-api/endpoint-reference/records/list-records
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	config := common.ReadParams{
		ObjectName: objectName,
		Fields:     datautils.NewSetFromList(fields),
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getObjectReadURL(objectName)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"filters": map[string]any{
			"record_id": map[string]any{
				"$in": ids,
			},
		},
	}

	res, err := c.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	parsed, err := common.ParseResult(res,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(url, config.ObjectName),
		DataMarshall(res),
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return parsed.Data, nil
}
