package zoho

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	//nolint:revive
	recordIds []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	if len(recordIds) == 0 {
		return nil, fmt.Errorf("%w: recordIds is empty", errMissingParams)
	}

	url, err := c.getAPIURL(crmAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("ids", strings.Join(recordIds, ","))

	fieldsNames := strings.Join(fields, ",")

	url.WithQueryParam("fields", fieldsNames)

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	parsed, err := common.ParseResult(res,
		common.ExtractRecordsFromPath("data"),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		datautils.NewSetFromList(fields),
	)
	if err != nil {
		return nil, err
	}

	return parsed.Data, nil
}
