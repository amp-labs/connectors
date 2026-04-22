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

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	suffix, err := getWriteSuffix(params)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName+suffix)
	if err != nil {
		return nil, err
	}

	body, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	if params.IsUpdate() {
		idKey := writeUpdateIdField[params.ObjectName]
		body[idKey] = params.RecordId
	}

	return jsonPostRequest(ctx, url.String(), body)
}

// Slack always returns HTTP 200, even on failure, so we inspect the "ok" field first.
//
//nolint:funlen,gocognit,cyclop,maintidx
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
		// If the Ok field is true but we don't have a spec for this object,
		// optimistically return success with no ID or data.
		// There are some write objects (e.g. reactions.add) that return no record data on success,
		return &common.WriteResult{Success: true}, nil
	}

	var recordNode *ajson.Node

	if spec.recordKey != "" {
		recordNode, err = jsonquery.New(body).ObjectRequired(spec.recordKey)
		if err != nil {
			return nil, err
		}
	} else {
		recordNode = body
	}

	recordID, err := jsonquery.New(recordNode).StrWithDefault(spec.idField, "")
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

func getWriteSuffix(params common.WriteParams) (string, error) {
	if params.IsUpdate() {
		updateSuffix, supported := writeUpdateSuffix[params.ObjectName]
		if !supported {
			return "", fmt.Errorf("%w: %s does not support updates",
				common.ErrOperationNotSupportedForObject, params.ObjectName)
		}

		return updateSuffix, nil
	}

	// Create: append ".add" or ".create" suffix.
	if writeObjectsUsingAddSuffix.Has(params.ObjectName) {
		return ".add", nil
	}

	return ".create", nil
}
