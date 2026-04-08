package slack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// buildWriteRequest constructs a POST request for the given Slack write object.
// All Slack write objects are create-only; updates are not supported.
// The API method suffix (".add" or ".create") is appended to the object name to form the URL path.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if params.RecordId != "" {
		return nil, fmt.Errorf("%w: %s does not support updates",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	suffix := ".create"
	if writeObjectsUsingAddSuffix.Has(params.ObjectName) {
		suffix = ".add"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName+suffix)
	if err != nil {
		return nil, err
	}

	body, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	return jsonPostRequest(ctx, url.String(), body)
}

// parseWriteResponse parses the Slack write response.
// Slack always returns HTTP 200, even on failure, so we inspect the "ok" field first.
func (c *Connector) parseWriteResponse(
	ctx context.Context, //nolint:revive
	params common.WriteParams,
	request *http.Request, //nolint:revive
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	okStatus, err := jsonquery.New(body).BoolRequired("ok")
	if err != nil {
		return nil, err
	}

	if !okStatus {
		// Map the Slack error code to a sentinel so callers can use errors.Is.
		errorCode, err := jsonquery.New(body).StringOptional("error")
		if err != nil {
			return nil, err
		}

		if errorCode != nil {
			return nil, interpretSlackErrorCode(*errorCode)
		}

		return nil, common.ErrBadProviderResponse
	}

	spec, found := writeResponseField[params.ObjectName]
	if !found || spec.idField == "" {
		// Objects that return no resource (e.g. reactions.add).
		return &common.WriteResult{Success: true}, nil
	}

	var recordNode *ajson.Node

	if spec.objectKey != "" {
		recordNode, err = jsonquery.New(body).ObjectRequired(spec.objectKey)
		if err != nil {
			return nil, err
		}
	} else {
		recordNode = body
	}

	recordID, err := jsonquery.New(recordNode).TextWithDefault(spec.idField, "")
	if err != nil {
		return nil, err
	}

	dataMap, err := jsonquery.Convertor.ObjectToMap(recordNode)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     dataMap,
	}, nil
}
