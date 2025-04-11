package zendeskchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	bans            = "bans"
	account         = "account"
	chats           = "chats"
	defaultPageSize = 1000
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.RootClient.URL(restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	firstRecord, err := parseIndividualResponse(objectName, response)
	if err != nil {
		return nil, err
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func parseIndividualResponse(objectName string, response *common.JSONHTTPResponse) (map[string]any, error) {
	field := responseField(objectName)
	if field != "" {
		resp, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, err
		}

		records, ok := (*resp)[field].([]any)
		if !ok {
			return nil, fmt.Errorf("expected field '%s' to contain an array, got %T: %w", field, (*resp)[field], common.ErrMissingExpectedValues) //nolint:lll
		}

		if len(records) == 0 {
			return nil, fmt.Errorf("no records found in response: %w", common.ErrMissingExpectedValues)
		}

		firstRecord, ok := records[0].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected first record in field '%s' to be a map, got %T: %w", field, records[0], common.ErrMissingExpectedValues) //nolint:lll
		}

		return firstRecord, nil
	}

	switch objectName {
	case bans, account:
		return parseObjectResponse(response)
	default:
		return parseListResponse(response)
	}
}

func parseObjectResponse(response *common.JSONHTTPResponse) (map[string]any, error) {
	resp, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, err
	}

	return *resp, nil
}

func parseListResponse(response *common.JSONHTTPResponse) (map[string]any, error) {
	resp, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, err
	}

	if len(*resp) == 0 {
		return nil, fmt.Errorf("no records found in response: %w", common.ErrMissingExpectedValues)
	}

	return (*resp)[0], nil
}

func (c *Connector) buildReadRequest(
	ctx context.Context,
	params common.ReadParams,
) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = c.RootClient.URL(restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseField(params.ObjectName)),
		nextRecordsURL(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(
	ctx context.Context,
	params common.WriteParams,
) (*http.Request, error) {
	method := http.MethodPost

	url, err := c.RootClient.URL(restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	// Endpoints routing_settings/account,routing_settings/agents/me,
	// and routing_settings/agents/{agent_id},
	// uses the PATCH method and these endpoints perform only updates.
	if _, exists := updatesByPatch[params.ObjectName]; exists {
		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Data:    data,
	}, nil
}
