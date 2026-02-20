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

	resp, err := c.executeRequest(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("error executing read request: %w", err)
	}

	bodyBytes := common.GetResponseBodyOnce(resp)

	var result map[string]any
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("error parsing read response: %w", err)
	}

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
		makeNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
