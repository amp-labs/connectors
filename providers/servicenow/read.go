package servicenow

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// Read wraps the component reader so ServiceNow's "no records" 404 becomes an empty
// page instead of an error. Several APIs return a 404 for an empty result set
// (including an incremental window with no changes) rather than an empty 200, which
// would otherwise fail every steady-state poll.
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	result, err := c.Reader.Read(ctx, params)
	if err != nil && isNoMatchingRecords(err) {
		return &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true}, nil
	}

	return result, err
}

// isNoMatchingRecords reports whether err is one of ServiceNow's 404 "empty result"
// responses (an empty result set rather than a true error):
//   - Lead API:    "No matching lead records found."
//   - TMF/Open API: "No Record found for given filter criteria" (code 60)
func isNoMatchingRecords(err error) bool {
	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusNotFound {
		return false
	}

	body := bytes.ToLower(httpErr.Body)

	return bytes.Contains(body, []byte("no matching")) ||
		bytes.Contains(body, []byte("no record found"))
}
