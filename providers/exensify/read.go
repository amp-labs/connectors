package exensify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

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
		// Expensify doesn't support pagination so makeNextRecordsURL always returns emptly string.
		makeNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
