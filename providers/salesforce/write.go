package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	// Delegated.
	return c.batchAdapter.BatchWrite(ctx, params)
}

// Write will write data to Salesforce.
//
//nolint:cyclop
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if c.isPardotModule() {
		return c.pardotAdapter.Write(ctx, config)
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

	headers := common.TransformWriteHeaders(config.Headers, common.HeaderModeOverwrite)

	rsp, err := c.Client.Post(ctx, url.String(), config.RecordData, headers...)
	if err != nil {
		return nil, err
	}

	rslt, err := parseWriteResult(rsp)
	if err != nil {
		return nil, err
	}

	if config.RecordId != "" && rslt.Success && rslt.RecordId == "" {
		rslt.RecordId = config.RecordId
	}

	return rslt, nil
}

// parseWriteResult parses the response from writing to Salesforce API. A 2xx return type is assumed.
func parseWriteResult(rsp *common.JSONHTTPResponse) (*common.WriteResult, error) {
	body, ok := rsp.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).StringRequired("id")
	if err != nil {
		return nil, err
	}

	errors, err := getErrors(body)
	if err != nil {
		return nil, err
	}

	success, err := jsonquery.New(body).BoolRequired("success")
	if err != nil {
		return nil, err
	}

	// Salesforce does not return record data upon successful write so we do not populate
	// the corresponding result field
	return &common.WriteResult{
		RecordId: recordID,
		Errors:   errors,
		Success:  success,
	}, nil
}

// getErrors returns the errors from the response.
func getErrors(node *ajson.Node) ([]any, error) {
	arr, err := jsonquery.New(node).ArrayOptional("errors")
	if err != nil {
		return nil, err
	}

	objects, err := jsonquery.Convertor.ArrayToObjects(arr)
	if err != nil {
		return nil, err
	}

	return objects, nil
}
