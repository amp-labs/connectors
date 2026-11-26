package salesloft

import (
	"context"
	"errors"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// GetRecordsByIds fetches records by id with the object's list endpoint and its ids[] filter.
//
// Ids are requested in groups of DefaultPageSize, the per_page value sent with each request, so every
// group fits on one page. When a group fails, each of its ids is reported in Failures, the rows of the
// other groups are still returned, and the errors are returned joined. A 404 still fails the whole call
// with its error. Requested ids that no response returned are reported as not-found failures.
//
// Reference: https://developers.salesloft.com/docs/api/account-stages-index/
//
//nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context, objectName string,
	recordIds []string, fields []string, _ []string,
) (*common.BatchReadResult, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName: objectName,
		Fields:     datautils.NewSetFromList(fields),
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	var (
		result = common.NewBatchReadResult(nil)
		// Ids of the groups whose request succeeded.
		fetchedIds []string
		errs       []error
	)

	for chunk := range slices.Chunk(recordIds, DefaultPageSize) {
		rows, err := c.getRecordsPage(ctx, config, chunk)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				return nil, err
			}

			for _, id := range chunk {
				result.AddFailure(id, err)
			}

			errs = append(errs, err)

			continue
		}

		result.Rows = append(result.Rows, rows...)
		fetchedIds = append(fetchedIds, chunk...)
	}

	result.ReportMissingIds(fetchedIds)

	return result, errors.Join(errs...)
}

// getRecordsPage fetches the first page of the object's list filtered to ids.
func (c *Connector) getRecordsPage(
	ctx context.Context, config common.ReadParams, ids []string,
) ([]common.ReadResultRow, error) {
	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	url.WithQueryParamList("ids[]", ids)

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
