package expensify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// Read is implemented manually rather than using the standard components.Reader
// because Expensify's response Content-Type is not application/json, so common.ParseJSONResponse
// rejects it, which means common.ParseResult (which depends on JSONHTTPResponse) cannot be used.
// We execute the request via the raw HTTPClient and build ReadResult directly from the parsed response map.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	err := config.ValidateParams(true)
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	body, err := buildReadBody(config.ObjectName)
	if err != nil {
		return nil, err
	}

	//nolint:bodyclose
	resp, err := c.executeRequest(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("error executing read request: %w", err)
	}

	bodyBytes := common.GetResponseBodyOnce(resp)

	var result map[string]any
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("error parsing read response: %w", err)
	}

	// by default Expensify returns 200 status code even for errors,
	//  so we need to check the response code in the body to determine if the request was successful or not
	if err = checkResponseCode(result); err != nil {
		return nil, err
	}

	jsonRes, err := common.ParseJSONResponse(resp, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing read response: %w", err)
	}

	return common.ParseResult(
		jsonRes,
		common.ExtractOptionalRecordsFromPath(readObjectResponseIdentifier.Get(config.ObjectName)),
		// Expensify doesn't support pagination so makeNextRecordsURL always returns empty string.
		makeNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
