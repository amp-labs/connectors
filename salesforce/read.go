package salesforce

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Read reads data from Salesforce. By default it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		location, joinErr := url.JoinPath(fmt.Sprintf("https://%s", c.Domain), config.NextPage)
		if joinErr != nil {
			return nil, joinErr
		}

		data, err = c.get(ctx, location)
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

		data, err = c.get(ctx, location+"?"+qp.Encode())
	}

	if err != nil {
		return nil, err
	}

	return parseResult(data)
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
		soql += fmt.Sprintf(" WHERE SystemModstamp > %s", config.Since.Format("2006-01-02T15:04:05Z"))
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

// parseResult parses the response from the Salesforce API. A 2xx return type is assumed.
func parseResult(data *ajson.Node) (*common.ReadResult, error) {
	totalSize, err := getTotalSize(data)
	if err != nil {
		return nil, err
	}

	records, err := getRecords(data)
	if err != nil {
		return nil, err
	}

	nextPage, err := getNextRecordsURL(data)
	if err != nil {
		return nil, err
	}

	done, err := getDone(data)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     totalSize,
		Data:     records,
		NextPage: common.NextPageToken(nextPage),
		Done:     done,
	}, nil
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

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("records")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, ErrNotArray
	}

	arr := records.MustArray()

	out := make([]map[string]interface{}, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, ErrNotObject
		}

		out = append(out, m)
	}

	return out, nil
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("nextRecordsUrl") {
		next, err := node.GetKey("nextRecordsUrl")
		if err != nil {
			return "", err
		}

		if !next.IsString() {
			return "", ErrNotString
		}

		nextPage = next.MustString()
	}

	return nextPage, nil
}

// getDone returns true if there are no more pages to read.
func getDone(node *ajson.Node) (bool, error) {
	var done bool

	if node.HasKey("done") {
		doneNode, err := node.GetKey("done")
		if err != nil {
			return false, err
		}

		if !doneNode.IsBool() {
			return false, ErrNotBool
		}

		done = doneNode.MustBool()
	}

	return done, nil
}

// getTotalSize returns the total number of records that match the query.
func getTotalSize(node *ajson.Node) (int64, error) {
	node, err := node.GetKey("totalSize")
	if err != nil {
		return 0, err
	}

	if !node.IsNumeric() {
		return 0, ErrNotNumeric
	}

	return int64(node.MustNumeric()), nil
}
