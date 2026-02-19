package exensify

import (
	"context"
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

	jsonRes, err := common.ParseJSONResponse(resp, common.GetResponseBodyOnce(resp))

	return common.ParseResult(
		jsonRes,
		common.ExtractOptionalRecordsFromPath("policyList"),
		makeNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
