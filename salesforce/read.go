package salesforce

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// Read reads data from Salesforce. By default it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		rsp *common.JSONHTTPResponse
		err error
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		location, joinErr := url.JoinPath("https://"+c.Domain, config.NextPage.String())
		if joinErr != nil {
			return nil, joinErr
		}

		rsp, err = c.get(ctx, location)
	} else {
		// If NextPage is not set, then we're reading the first page of results.
		// We need to construct the SOQL query and then make the request.
		soql, soqlErr := makeSOQL(config)
		if soqlErr != nil {
			return nil, soqlErr
		}

		// Encode the SOQL query as a URL parameter
		qp := url.Values{}
		qp.Add("q", soql)

		location, joinErr := url.JoinPath(c.BaseURL, "/query/")
		if joinErr != nil {
			return nil, joinErr
		}

		rsp, err = c.get(ctx, location+"?"+qp.Encode())
	}

	if err != nil {
		if strings.Contains(err.Error(), "invalid_grant") {
			return nil, common.ErrInvalidGrant
		}
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		config.Fields,
	)
}

// makeSOQL returns the SOQL query for the desired read operation.
func makeSOQL(config common.ReadParams) (string, error) {
	// Make sure we have at least one field
	if len(config.Fields) == 0 {
		return "", ErrNoFields
	}

	// Get the field set in SOQL format
	fields := getFieldSet(config.Fields)

	hasWhere := false
	soql := fmt.Sprintf("SELECT %s FROM %s", fields, config.ObjectName)

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql += " WHERE SystemModstamp > " + config.Since.Format("2006-01-02T15:04:05Z")
		hasWhere = true
	}

	if config.Deleted {
		if !hasWhere {
			soql += " WHERE"
		} else {
			soql += " AND"
		}

		soql += " IsDeleted = true"
	}

	return soql, nil
}

// getFieldSet returns the field set in SOQL format.
func getFieldSet(fields []string) string {
	for _, field := range fields {
		if field == "*" {
			return "FIELDS(ALL)"
		}
	}

	return strings.Join(fields, ",")
}
