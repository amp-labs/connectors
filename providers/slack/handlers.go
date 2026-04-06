package slack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := objectName
	if !objectsWithoutListSuffix.Has(objectName) {
		urlPath = objectName + ".list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlPath)
	if err != nil {
		return nil, err
	}

	if postMethodObjects.Has(objectName) {
		return jsonPostRequest(ctx, url.String(), map[string]any{"limit": 1})
	}

	url.WithQueryParam("limit", "1")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

//nolint:cyclop,funlen
func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateDotWords(objectName)),
	}

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// Slack always returns HTTP 200, even for errors. The actual success or failure
	// is indicated by the "ok" field in the response body. We check it here so that
	// metadata calls fail clearly instead of trying to parse a response that has no records.
	okStatus, okCast := (*res)["ok"].(bool)
	if !okCast {
		return nil, fmt.Errorf("couldn't cast 'ok' field to bool: %w", common.ErrMissingExpectedValues)
	}

	if !okStatus {
		// When ok is false, Slack usually includes an "error" field with a short error code.
		// Map it to a sentinel error so callers can use errors.Is to react appropriately.
		errorCode, ok := (*res)["error"].(string)
		if ok {
			return nil, interpretSlackErrorCode(errorCode)
		}

		return nil, fmt.Errorf("failed response: %w", common.ErrMissingExpectedValues)
	}

	responseKey := objectResponseField.Get(objectName)

	responseValue, exists := (*res)[responseKey]
	if !exists {
		return nil, fmt.Errorf("response key %q not found: %w", responseKey, common.ErrMissingExpectedValues)
	}

	records, ok := responseValue.([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert response field %q to an array: %w", responseKey, common.ErrMissingExpectedValues) //nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "",
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	urlPath := params.ObjectName
	if !objectsWithoutListSuffix.Has(params.ObjectName) {
		urlPath = params.ObjectName + ".list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlPath)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, "200")

	if postMethodObjects.Has(params.ObjectName) {
		body := map[string]any{"limit": pageSize}
		if params.NextPage != "" {
			body["cursor"] = params.NextPage.String()
		}

		return jsonPostRequest(ctx, url.String(), body)
	}

	url.WithQueryParam("limit", pageSize)

	if params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse( //nolint:unparam
	ctx context.Context, //nolint:revive
	params common.ReadParams,
	request *http.Request, //nolint:revive
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
