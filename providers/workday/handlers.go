package workday

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// defaultPageLimit is the default number of records per page for Workday API requests.
	defaultPageLimit = "100"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "ccx", "api", "v1", c.tenantName, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageLimit))

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// parseReadResponse parses the Workday API response.
//
// Workday REST API v1 response format:
//
//	{
//	  "data": [ ... ],   // array of record objects
//	  "total": 100       // total number of records available
//	}
//
// Records are extracted from the "data" array. Pagination is offset-based,
// computed by comparing the current offset + returned records against "total".
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	registry, err := c.fetchCustomFieldDefinitions(ctx, params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.MakeRecordsFunc("data"), // records live in the "data" array
		makeNextRecordsURL(reqURL),
		common.MakeMarshaledDataFunc(c.attachReadCustomFields(registry)),
		params.Fields,
	)
}

// makeNextRecordsURL creates a NextPageFunc that computes the next page URL
// using offset-based pagination. Workday returns a `total` count in the response;
// if the current offset + number of records returned is less than total,
// the next URL is returned with an incremented offset.
// An empty string signals that all records have been fetched.
func makeNextRecordsURL(reqURL *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		total, err := jsonquery.New(node).IntegerWithDefault("total", 0)
		if err != nil {
			return "", err
		}

		records, err := jsonquery.New(node).ArrayOptional("data")
		if err != nil {
			return "", err
		}

		numRecords := int64(len(records))
		if numRecords == 0 {
			return "", nil
		}

		// Determine current offset from the request URL.
		offsetText, _ := reqURL.GetFirstQueryParam("offset")
		if offsetText == "" {
			offsetText = "0"
		}

		offset, err := strconv.ParseInt(offsetText, 10, 64)
		if err != nil {
			return "", err
		}

		nextOffset := offset + numRecords
		if nextOffset >= total {
			return "", nil
		}

		reqURL.WithQueryParam("offset", strconv.FormatInt(nextOffset, 10))

		return reqURL.String(), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "ccx", "api", "v1", c.tenantName, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

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
	body, ok := response.Body()
	if !ok {
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
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL, "ccx", "api", "v1", c.tenantName, params.ObjectName, params.RecordId,
	)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return &common.DeleteResult{
		Success: true,
	}, nil
}

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	if res.StatusCode == http.StatusNotFound {
		return common.NewHTTPError(res.StatusCode, body, common.GetResponseHeaders(res), common.ErrNotFound)
	}

	return common.InterpretError(res, body)
}
