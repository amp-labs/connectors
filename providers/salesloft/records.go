package salesloft

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Reference: https://developers.salesloft.com/docs/api/account-stages-index/
//
//nolint:godoclint,revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	params common.ReadByIdsParams,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName: params.ObjectName,
		Fields:     datautils.NewSetFromList(params.Fields),
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	url.WithQueryParamList("ids[]", params.RecordIds)

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	parsed, err := common.ParseResult(rsp,
		getRecords,
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return parsed.Data, nil
}
