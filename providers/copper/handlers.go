package copper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/copper/internal/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 200

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method, jsonData, err := createReadOperation(url, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	applicationHeader.ApplyToRequest(req)
	c.emailHeader().ApplyToRequest(req)

	return req, nil
}

// createReadOperation such that READ operation is either:
// * POST with a payload for /search based URLs.
// * GET without a payload for the rest of URLs.
func createReadOperation(
	url *urlbuilder.URL, params common.ReadParams,
) (method string, payload []byte, err error) {
	if !strings.HasSuffix(url.Path(), "/search") {
		return http.MethodGet, nil, nil
	}

	searchPayload := map[string]any{
		"page_number":    getPageNum(params),
		"page_size":      defaultPageSize,
		"sort_by":        "date_modified",
		"sort_direction": "desc",
	}

	if strings.HasSuffix(url.Path(), "/users/search") {
		// Users doesn't have incremental read.
	} else {
		if !params.Since.IsZero() {
			searchPayload["minimum_modified_date"] = strconv.FormatInt(params.Since.Unix(), 10)
		}

		if !params.Until.IsZero() {
			searchPayload["maximum_modified_date"] = strconv.FormatInt(params.Until.Unix(), 10)
		}
	}

	payload, err = json.Marshal(searchPayload)
	if err != nil {
		return "", nil, err
	}

	return http.MethodPost, payload, nil
}

func getPageNum(params common.ReadParams) string {
	if nextPage := params.NextPage.String(); nextPage != "" {
		return nextPage
	}

	return "1"
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	customFields, err := c.fetchCustomFields(ctx)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		makeGetRecords(params.ObjectName),
		makeNextRecordsURL(params),
		common.MakeMarshaledDataFunc(c.attachReadCustomFields(customFields)),
		params.Fields,
	)
}

func makeGetRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

// https://developer.copper.com/introduction/pagination.html#strategy-2-count-the-records-on-each-page
func makeNextRecordsURL(params common.ReadParams) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		nextPage := params.NextPage.String()
		if nextPage == "" {
			// Default to the second page, the first page was already read.
			return "2", nil
		}

		pageNum, err := strconv.Atoi(nextPage)
		if err != nil {
			return "", err
		}

		pageNum += 1

		return strconv.Itoa(pageNum), nil
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

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	applicationHeader.ApplyToRequest(req)
	c.emailHeader().ApplyToRequest(req)

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

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
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

	applicationHeader.ApplyToRequest(req)
	c.emailHeader().ApplyToRequest(req)

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
