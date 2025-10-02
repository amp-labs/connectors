package salesflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// Documentation has no mention of the upper limit.
	defaultPageSize = "10000"
	envPageSize     = "SALESFLARE_PAGE_SIZE"
)

var pageSize = defaultPageSize // nolint:gochecknoglobals

func init() {
	if value, ok := os.LookupEnv(envPageSize); ok {
		if number, err := strconv.Atoi(value); err == nil && number > 0 {
			pageSize = value
		}
	}
}

func (c Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.getReadURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c Connector) parseSingleHandlerResponse(
	ctx context.Context, objectName string, request *http.Request, response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	fields := make(common.FieldsMetadata)

	records, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, err
	}

	if records == nil || len(*records) < 1 {
		return nil, common.ErrMissingFields
	}

	for fieldName := range (*records)[0] {
		fields.AddFieldWithDisplayOnly(fieldName, naming.CapitalizeFirstLetterEveryWord(fieldName))
	}

	return common.NewObjectMetadata(
		naming.CapitalizeFirstLetterEveryWord(objectName), fields,
	), nil
}

// nolint:gochecknoglobals
var (
	objectNameToSinceQueryParameter = map[string]string{
		"accounts":      "creation_after",
		"contacts":      "modification_after",
		"me/contacts":   "modification_after",
		"opportunities": "creation_after",
	}
	objectNameToUntilQueryParameter = map[string]string{
		"accounts":      "creation_before",
		"contacts":      "modification_before",
		"me/contacts":   "modification_before",
		"opportunities": "creation_before",
	}
)

func (c Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", pageSize)

	if !params.Since.IsZero() {
		if query := objectNameToSinceQueryParameter[params.ObjectName]; query != "" {
			url.WithQueryParam(query, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
		}
	}

	if !params.Until.IsZero() {
		if query := objectNameToUntilQueryParameter[params.ObjectName]; query != "" {
			url.WithQueryParam(query, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Until))
		}
	}

	return url, nil
}

func (c Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(""),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

// makeNextRecordsURL creates a NextPageFunc that advances pagination by
// increasing the "offset" query parameter based on the number of records
// returned in the current page.
func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		records, err := jsonquery.New(node).ArrayOptional("")
		if err != nil {
			return "", err
		}

		numRecords := int64(len(records))

		if numRecords == 0 {
			return "", nil
		}

		offsetText, _ := url.GetFirstQueryParam("offset")

		if offsetText == "" {
			offsetText = "0"
		}

		offset, err := strconv.ParseInt(offsetText, 10, 64)
		if err != nil {
			return "", err
		}

		offset += numRecords

		url.WithQueryParam("offset", strconv.FormatInt(offset, 10))

		return url.String(), nil
	}
}

func (c Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if len(params.RecordId) != 0 {
		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	// Provider API does not return entity. Response only includes the id of created instance.
	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}

func (c Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getDeleteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}
