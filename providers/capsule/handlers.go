package capsule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
	"github.com/spyzhov/ajson"
)

// DefaultPageSize
// https://developer.capsulecrm.com/v2/overview/reading-from-the-api
const DefaultPageSize = "100"

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

	url.WithQueryParam("perPage", DefaultPageSize)

	if !params.Since.IsZero() {
		url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	if len(params.AssociatedObjects) != 0 {
		url.WithQueryParam("embed", strings.Join(params.AssociatedObjects, ","))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developer.capsulecrm.com/v2/overview/reading-from-the-api
func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if len(params.RecordId) != 0 {
		method = http.MethodPut
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// Wrap user payload in the named object required by the API.
	nestedKey := nestedWriteObject(params.ObjectName)
	payload := map[string]any{
		nestedKey: recordData,
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

	nestedKey := nestedWriteObject(params.ObjectName)

	nested, err := jsonquery.New(body).ObjectRequired(nestedKey)
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
