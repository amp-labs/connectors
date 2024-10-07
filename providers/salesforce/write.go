package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write will write data to Salesforce.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getRestApiURL("sobjects", config.ObjectName)
	if err != nil {
		return nil, err
	}

	if config.RecordId != "" {
		url.AddPath(config.RecordId)
		// Salesforce allows for PATCH method override
		url.WithQueryParam("_HttpMethod", "PATCH")
	}

	rsp, err := c.JSON.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return parseWriteResult(rsp)
}

// parseWriteResult parses the response from writing to Salesforce API. A 2xx return type is assumed.
func parseWriteResult(rsp *common.JSONHTTPResponse) (*common.WriteResult, error) {
	body, ok := rsp.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).Str("id", false)
	if err != nil {
		return nil, err
	}

	errors, err := getErrors(body)
	if err != nil {
		return nil, err
	}

	success, err := jsonquery.New(body).Bool("success", false)
	if err != nil {
		return nil, err
	}

	// Salesforce does not return record data upon successful write so we do not populate
	// the corresponding result field
	return &common.WriteResult{
		RecordId: *recordID,
		Errors:   errors,
		Success:  *success,
	}, nil
}

// getErrors returns the errors from the response.
func getErrors(node *ajson.Node) ([]any, error) {
	arr, err := jsonquery.New(node).Array("errors", true)
	if err != nil {
		return nil, err
	}

	objects, err := jsonquery.Convertor.ArrayToObjects(arr)
	if err != nil {
		return nil, err
	}

	return objects, nil
}
