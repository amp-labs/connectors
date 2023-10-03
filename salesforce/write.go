package salesforce

import (
	"context"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Writes data to Salesforce
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	location, joinErr := url.JoinPath(fmt.Sprintf("%s/sobjects", c.BaseURL), config.ObjectName)
	if joinErr != nil {
		return nil, joinErr
	}

	if config.ObjectId != "" {
		location, joinErr = url.JoinPath(location, config.ObjectId)
		if joinErr != nil {
			return nil, joinErr
		}
		// Salesforce allows for PATCH method override
		location += "?_HttpMethod=PATCH"
	}
	data, err = c.post(ctx, location, config.ObjectData)

	if err != nil {
		return nil, err
	}

	return parseWriteResult(data)
}

// parseWriteResult parses the response from writing to Salesforce API. A 2xx return type is assumed.
func parseWriteResult(data *ajson.Node) (*common.WriteResult, error) {

	// in case we got a 204 and empty array => unmarshal into nil ajson node
	if data == nil {
		return nil, nil
	}

	createdObjectId, err := getCreatedObjectId(data)
	if err != nil {
		return nil, err
	}

	errors, err := getErrors(data)
	if err != nil {
		return nil, err
	}

	success, err := getSuccess(data)
	if err != nil {
		return nil, err
	}
	return &common.WriteResult{
		RespData: map[string]interface{}{
			"id":      createdObjectId,
			"errors":  errors,
			"success": success,
		},
	}, nil
}

// getErrors returns the errors from the response.
// OKADA TODO: are errors always strings
func getErrors(node *ajson.Node) ([]string, error) {
	errors, err := node.GetKey("errors")
	if err != nil {
		return nil, err
	}

	if !errors.IsArray() {
		return nil, ErrNotArray
	}

	arr := errors.MustArray()

	out := make([]string, 0, len(arr))

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

func getCreatedObjectId(node *ajson.Node) (string, error) {
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
