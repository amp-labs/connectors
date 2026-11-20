package zoho

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
)

// crmMaxPageSize is the largest per_page Zoho CRM accepts when listing records.
// https://www.zoho.com/crm/developer/docs/api/v6/get-records.html
const crmMaxPageSize = 200

// GetRecordsByIds fetches records by id. Zoho Mail delegates to its adapter; the other modules list the
// object with the ids filter.
//
// Ids are requested in groups of crmMaxPageSize with per_page set to the same value, so every group
// fits on one page. When a group fails, each of its ids is reported in Failures, the rows of the other
// groups are still returned, and the errors are returned joined. Requested ids that no response
// returned are reported as not-found failures.
//
//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	//nolint:revive
	recordIds []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	// The Mail module fetches records through its own endpoints, addressed by
	// composite ids ("<folderId>/<messageId>" for messages, "<groupId>/<taskId>"
	// for group tasks), so delegate to the mail adapter.
	if c.moduleID == providers.ModuleZohoMail {
		return c.mailAdapter.GetRecordsByIds(ctx, objectName, recordIds, fields, associations)
	}

	if len(recordIds) == 0 {
		return nil, fmt.Errorf("%w: recordIds is empty", errMissingParams)
	}

	var (
		result = common.NewBatchReadResult(nil)
		// Ids of the groups whose request succeeded.
		fetchedIds []string
		errs       []error
	)

	for chunk := range slices.Chunk(recordIds, crmMaxPageSize) {
		rows, err := c.getRecordsPage(ctx, objectName, chunk, fields)
		if err != nil {
			// Zoho documents 404 as "Invalid request", which is about the request rather than the ids,
			// so it fails the whole call. The 404 status is left out of the returned error so that it
			// is not read as the records being missing.
			// https://www.zoho.com/crm/developer/docs/api/v6/status-codes.html
			if httpErr, ok := errors.AsType[*common.HTTPError](err); ok && httpErr.Status == http.StatusNotFound {
				return nil, fmt.Errorf("%w: %v", common.ErrBadRequest, err) //nolint:errorlint
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

// getRecordsPage fetches the first page of the object's records filtered to ids.
func (c *Connector) getRecordsPage(
	ctx context.Context, objectName string, ids []string, fields []string,
) ([]common.ReadResultRow, error) {
	url, err := c.getAPIURL(crmAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("ids", strings.Join(ids, ","))
	url.WithQueryParam("fields", strings.Join(fields, ","))
	url.WithQueryParam("per_page", strconv.Itoa(crmMaxPageSize))

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	// Zoho answers 204 with no body when "There is no content available for the request". No rows are
	// returned, so the requested ids are reported as not found.
	// https://www.zoho.com/crm/developer/docs/api/v6/status-codes.html
	if _, hasData := res.Body(); !hasData {
		return nil, nil
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
