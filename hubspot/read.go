package hubspot

import (
	"context"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Read reads data from Hubspot. By default, it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		data, err = c.get(ctx, config.NextPage)
	} else {
		// If NextPage is not set, then we're reading the first page of results.
		// We need to construct the SOQL query and then make the request.
		data, err = c.get(ctx, c.BaseURL+"/"+config.ObjectName+"?"+makeQuery(config))
	}

	if err != nil {
		return nil, err
	}

	return parseResult(data)
}

// makeQuery returns the query for the desired read operation.
func makeQuery(config common.ReadParams) string {
	queryValues := url.Values{}

	if len(config.Fields) != 0 {
		queryValues.Add("properties", strings.Join(config.Fields, ","))
	}

	if config.Deleted {
		queryValues.Add("archived", "true")
	}

	return queryValues.Encode()
}

// parseResult parses the response from the Hubspot API. A 2xx return type is assumed.
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

	done := nextPage == ""

	return &common.ReadResult{
		Rows:     totalSize,
		Data:     records,
		NextPage: common.NextPageToken(nextPage),
		Done:     done,
	}, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("results")
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
	// No paging key signifies that there are no more results.
	if !node.HasKey("paging") {
		return "", nil
	}

	paging, err := node.GetKey("paging")
	if err != nil {
		return "", err
	}

	if !paging.IsObject() {
		return "", ErrNotObject
	}

	next, err := paging.GetKey("next")
	if err != nil {
		return "", err
	}

	if !next.IsObject() {
		return "", ErrNotObject
	}

	link, err := next.GetKey("link")
	if err != nil {
		return "", err
	}

	if !link.IsString() {
		return "", ErrNotString
	}

	return link.MustString(), nil
}

// getTotalSize returns the total number of records that match the query.
func getTotalSize(node *ajson.Node) (int64, error) {
	node, err := node.GetKey("results")
	if err != nil {
		return 0, err
	}

	if !node.IsArray() {
		return 0, ErrNotArray
	}

	return int64(node.Size()), nil
}
