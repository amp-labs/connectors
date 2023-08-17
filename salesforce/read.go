package salesforce

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Read reads data from Salesforce. By default it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (s *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var data *ajson.Node
	var err error

	// Make sure we have at least one field
	if len(config.Fields) == 0 {
		return nil, errors.New("no fields specified")
	}

	// Get the field set in SOQL format
	fields := getFieldSet(config.Fields)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results. All that matters is the URL, the fields are ignored.
		data, err = s.get(ctx, fmt.Sprintf("https://%s%s", s.Domain, config.NextPage))
	} else if config.Since.IsZero() {
		// If Since is not set, then we're doing a backfill. We read all rows (in pages)
		soql := fmt.Sprintf("SELECT %s FROM %s", fields, config.ObjectName)

		qp := url.Values{}
		qp.Add("q", soql)
		data, err = s.get(ctx, s.BaseURL+"/query/?"+qp.Encode())
	} else {
		// If Since is set, then we're reading only rows that have been updated since the specified time.
		soql := fmt.Sprintf("SELECT %s FROM %s WHERE SystemModstamp > %s", fields, config.ObjectName, config.Since.Format("2006-01-02T15:04:05Z"))

		qp := url.Values{}
		qp.Add("q", soql)
		data, err = s.get(ctx, s.BaseURL+"/query/?"+qp.Encode())
	}

	if err != nil {
		return nil, err
	}

	ts, err := getTotalSize(data)
	if err != nil {
		return nil, err
	}

	records, err := getRecords(data)
	if err != nil {
		return nil, err
	}

	nextPage, err := getNextRecordsUrl(data)
	if err != nil {
		return nil, err
	}

	done, err := getDone(data)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     ts,
		Data:     records,
		NextPage: nextPage,
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
		return nil, errors.New("records isn't an array")
	}

	arr := records.MustArray()
	var out []map[string]interface{}

	for _, v := range arr {
		if !v.IsObject() {
			return nil, errors.New("record isn't an object")
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, errors.New("record isn't an object")
		}

		out = append(out, m)
	}

	return out, nil
}

// getNextRecordsUrl returns the URL for the next page of results.
func getNextRecordsUrl(node *ajson.Node) (string, error) {
	var nextPage string
	if node.HasKey("nextRecordsUrl") {
		next, err := node.GetKey("nextRecordsUrl")
		if err != nil {
			return "", err
		}

		if !next.IsString() {
			return "", errors.New("nextRecordsUrl isn't a string")
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
			return false, errors.New("done isn't a boolean")
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
		return 0, errors.New("totalSize isn't numeric")
	}

	return int64(node.MustNumeric()), nil
}
