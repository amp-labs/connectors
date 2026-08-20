package slack

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/slack/internal/mappings"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		resourceName string
		idKey        string
	)

	if params.IsUpdate() {
		info, err := mappings.GetWriteUpdateInfo(c.Provider(), params.ObjectName)
		if err != nil {
			return nil, err
		}

		resourceName = info.Href
		idKey = info.RequestIdField
	} else {
		info, err := mappings.GetWriteCreateInfo(c.Provider(), params.ObjectName)
		if err != nil {
			return nil, err
		}

		resourceName = info.Href
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, resourceName)
	if err != nil {
		return nil, err
	}

	body, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	if params.IsUpdate() {
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

	var (
		recordKey string
		idField   string
	)

	if params.IsUpdate() {
		info, err := mappings.GetWriteUpdateInfo(c.Provider(), params.ObjectName)
		if err != nil {
			return nil, err
		}

		recordKey = info.ResponseField
		idField = info.ResponseIdField
	} else {
		info, err := mappings.GetWriteCreateInfo(c.Provider(), params.ObjectName)
		if err != nil {
			return nil, err
		}

		recordKey = info.ResponseField
		idField = info.ResponseIdField
	}

	var recordNode *ajson.Node
	if recordKey != "" {
		recordNode, err = jsonquery.New(body).ObjectRequired(recordKey)
		if err != nil {
			return nil, err
		}
	} else {
		recordNode = body
	}

	recordID := params.RecordId
	if idField != "" {
		// Most responses contain a field that identifies the affected record.
		// Some successful operations return only a status response without a
		// meaningful record ID. In that case, fall back to the ID from the params.
		recordID, err = jsonquery.New(recordNode).StrWithDefault(idField, params.RecordId)
		if err != nil {
			return nil, err
		}
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
