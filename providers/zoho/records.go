package zoho

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	params common.ReadByIdsParams,
) ([]common.ReadResultRow, error) {
	if len(params.RecordIds) == 0 {
		return nil, fmt.Errorf("%w: recordIds is empty", errMissingParams)
	}

	url, err := c.getAPIURL(crmAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("ids", strings.Join(params.RecordIds, ","))
	url.WithQueryParam("fields", strings.Join(params.Fields, ","))

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	parsed, err := common.ParseResult(res,
		common.ExtractRecordsFromPath("data"),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		datautils.NewSetFromList(params.Fields),
	)
	if err != nil {
		return nil, err
	}

	return parsed.Data, nil
}
