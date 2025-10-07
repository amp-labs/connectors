package teamwork

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/teamwork/internal/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "500"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", defaultPageSize)

	if !params.Since.IsZero() {
		url.WithQueryParam("updatedAfter", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageObject := jsonquery.New(node, "meta", "page")

		hasMore, err := pageObject.BoolWithDefault("hasMore", false)
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		pageOffset, err := pageObject.IntegerOptional("pageOffset")
		if err != nil {
			return "", err
		}

		if pageOffset == nil {
			return "", nil
		}

		nextPageOffset := *pageOffset + 1
		url.WithQueryParam("page", strconv.FormatInt(nextPageOffset, 10))

		return url.String(), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if len(params.RecordId) != 0 {
		method = http.MethodPatch
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// Payload is nested under the key matching object name in singular case.
	payloadName := writeObjectName(params.ObjectName)
	payload := map[string]any{
		payloadName: recordData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	writeResponseFieldName := writeObjectName(params.ObjectName)

	nested, err := jsonquery.New(body).ObjectRequired(writeResponseFieldName)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(nested).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(nested)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
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

// writeObjectName is used by payload and response.
// This field is object name in singular case.
func writeObjectName(objectName string) string {
	readResponseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)
	writeResponseFieldName := naming.NewSingularString(readResponseFieldName).String()

	return writeResponseFieldName
}
