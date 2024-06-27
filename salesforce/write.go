package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Write will write data to Salesforce.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	url, err := c.getURL("sobjects", config.ObjectName)
	if err != nil {
		return nil, err
	}

	if config.RecordId != "" {
		url.AddPath(config.RecordId)
		// Salesforce allows for PATCH method override
		url.WithQueryParam("_HttpMethod", "PATCH")
	}

	rsp, err := c.Client.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return parseWriteResult(rsp)
}

// parseWriteResult parses the response from writing to Salesforce API. A 2xx return type is assumed.
func parseWriteResult(rsp *common.JSONHTTPResponse) (*common.WriteResult, error) {
	// in case we got a 204 and empty array => unmarshal into nil ajson node
	if rsp == nil || rsp.Body == nil {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	createdRecordId, err := getCreatedRecordId(rsp.Body)
	if err != nil {
		return nil, err
	}

	errors, err := getErrors(rsp.Body)
	if err != nil {
		return nil, err
	}

	success, err := getSuccess(rsp.Body)
	if err != nil {
		return nil, err
	}

	// Salesforce does not return record data upon successful write so we do not populate
	// the corresponding result field
	return &common.WriteResult{
		RecordId: createdRecordId,
		Errors:   errors,
		Success:  success,
	}, nil
}

// getErrors returns the errors from the response.
func getErrors(node *ajson.Node) ([]any, error) {
	errors, err := node.GetKey("errors")
	if err != nil {
		return nil, err
	}

	if !errors.IsArray() {
		return nil, ErrNotArray
	}

	arr := errors.MustArray()

	out := make([]any, 0, len(arr))

	for _, v := range arr {
		if !v.IsString() {
			return nil, ErrNotString
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(string)
		if !ok {
			return nil, ErrNotString
		}

		out = append(out, m)
	}

	return out, nil
}

func getCreatedRecordId(node *ajson.Node) (string, error) {
	idNode, err := node.GetKey("id")
	if err != nil {
		return "", err
	}

	return idNode.MustString(), nil
}

func getSuccess(node *ajson.Node) (bool, error) {
	successNode, err := node.GetKey("success")
	if err != nil {
		return false, err
	}

	if !successNode.IsBool() {
		return false, ErrNotBool
	}

	return successNode.MustBool(), nil
}
