package salesloft

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Reference: https://developers.salesloft.com/docs/api/account-stages-index/
//
//nolint:revive, godoclint
func (c *Connector) GetRecordsByIds(ctx context.Context, objectName string,
	recordIds []string, fields []string, _ []string,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName: objectName,
		Fields:     datautils.NewSetFromList(fields),
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("ids", strings.Join(recordIds, ","))

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
