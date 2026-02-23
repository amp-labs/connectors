package expensify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	body, err := buildWriteBody(params.RecordData)
	if err != nil {
		return nil, err
	}

	//nolint:bodyclose
	resp, err := c.executeRequest(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("error executing write request: %w", err)
	}

	bodyBytes := common.GetResponseBodyOnce(resp)

	var result map[string]any
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("error parsing write response: %w", err)
	}

	// by default Expensify returns 200 status code even for errors,
	//  so we need to check the response code in the body to determine if the request was successful or not
	if err = checkResponseCode(result); err != nil {
		return nil, err
	}

	responseIdKey := params.ObjectName + "ID"

	recordID, _ := result[responseIdKey].(string)

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     result,
	}, nil
}
