package claricopilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	limitQuery       = "limit"
	metadataPageSize = "1"
	pageSize         = "100"
	skipKey          = "skip"
	apiVersionV2     = "v2"
	createPrefix     = "create"
	updatePrefix     = "update"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if supportedObjectV2.Has(objectName) {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersionV2, objectName)
	} else {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	url.WithQueryParam(limitQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.NewObjectMetadata(
		naming.CapitalizeFirstLetterEveryWord(objectName),
		common.FieldsMetadata{},
	)

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	records, ok := (*data)[objectName].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field := range firstRecord {
		objectMetadata.AddField(field, field)
	}

	return objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if supportedObjectV2.Has(params.ObjectName) {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersionV2, params.ObjectName)
	} else {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("filterModifiedGt", params.Since.Format(time.RFC3339))
	}

	url.WithQueryParam(limitQuery, pageSize)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, fmt.Errorf("failed to build URL from next page: %w", err)
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
		nextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	objectName := writeObjectMapping.Get(params.ObjectName)

	fullObjectName := fmt.Sprintf("%s-%s", createPrefix, objectName)

	if params.RecordId != "" {
		method = http.MethodPut
		fullObjectName = fmt.Sprintf("%s-%s", updatePrefix, objectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, fullObjectName)
	if err != nil {
		return nil, err
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
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("crm_id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     resp,
	}, nil
}
