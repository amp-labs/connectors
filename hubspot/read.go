package hubspot

import (
	"context"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Read reads data from Hubspot. If Since is set, it will use the
// search endpoint to filter records.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	// If filtering is required, then we have to use the search endpoint.
	if requiresFiltering(config) {
		return nil, ErrFilteringNotSupported
	}

	// Object endpoints have a link

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		data, err = c.get(ctx, config.NextPage)
	} else {
		// If NextPage is not set, then we're reading the first page of results.
		// We need to construct the SOQL query and then make the request.
		data, err = c.get(ctx, c.BaseURL+"/objects/"+config.ObjectName+"?"+makeQueryValues(config))
	}

	if err != nil {
		return nil, err
	}

	return parseResult(data, getNextRecordsURL)
}

// makeQueryValues returns the query for the desired read operation.
func makeQueryValues(config common.ReadParams) string {
	queryValues := url.Values{}

	if len(config.Fields) != 0 {
		queryValues.Add("properties", strings.Join(config.Fields, ","))
	}

	if config.Deleted {
		queryValues.Add("archived", "true")
	}

	queryValues.Add("limit", DefaultPageSize)

	return queryValues.Encode()
}

// parseResult parses the response from the Hubspot API. A 2xx return type is assumed.
func parseResult(data *ajson.Node, paginationFunc func(*ajson.Node) (string, error)) (*common.ReadResult, error) {
	totalSize, err := getTotalSize(data)
	if err != nil {
		return nil, err
	}

	records, err := getRecords(data)
	if err != nil {
		return nil, err
	}

	nextPage, err := paginationFunc(data)
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
