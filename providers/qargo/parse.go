package qargo

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Last update time of a company, an ISO 8601 UTC datetime.
// https://api-docs.qargo.com/docs/api-company/list_companies_v1_companies_company_get
const updatedField = "timestamp_updated"

// makeUpdatedFilter keeps the records updated within [Since, Until] and follows next_cursor to the end.
// The docs give no order for the list, so every page is read. A page may be empty while
// next_cursor is still set, so paging continues on the cursor alone.
// https://api-docs.qargo.com/docs/section/api-best-practices/pagination
func makeUpdatedFilter() common.RecordsFilterFunc {
	boundary := readhelper.NewTimeBoundary()

	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		next, err := nextCursor(body)
		if err != nil {
			return nil, "", err
		}

		if params.Since.IsZero() && params.Until.IsZero() {
			return records, next, nil
		}

		filtered := make([]*ajson.Node, 0, len(records))

		for _, record := range records {
			updated, err := lastUpdated(record)
			if err != nil {
				return nil, "", err
			}

			if boundary.Contains(params, updated) {
				filtered = append(filtered, record)
			}
		}

		return filtered, next, nil
	}
}

func nextCursor(body *ajson.Node) (string, error) {
	cursor, err := jsonquery.New(body).StringOptional("next_cursor")
	if err != nil || cursor == nil {
		return "", err
	}

	return *cursor, nil
}

func lastUpdated(record *ajson.Node) (time.Time, error) {
	value, err := jsonquery.New(record).StringRequired(updatedField)
	if err != nil {
		return time.Time{}, err
	}

	updated, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s %q", common.ErrMissingExpectedValues, updatedField, value)
	}

	return updated, nil
}
