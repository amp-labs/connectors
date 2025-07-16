package flatfile

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// nolint: gochecknoglobals
var (
	version       = "v1"
	pageSize      = "100"
	pageSizeQuery = "pageSize"
	pageQuery     = "pageNumber"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, version, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageSizeQuery, pageSize)
	url.WithQueryParam(pageQuery, "1") // Start with the first page

	if supportObjectSince.Has(params.ObjectName) && !params.Since.IsZero() {
		url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTC(params.Since))
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
		records(),
		nextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) { //nolint:lll
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, version, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
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

	dataNode, err := jsonquery.New(body).ObjectOptional("data")
	if err != nil {
		return nil, err
	}

	if dataNode == nil {
		// If the "data" field does not exist, use the root body as the data node.
		dataNode = body
	}

	recordID, err := jsonquery.New(dataNode).StringRequired("id")
	if err != nil {
		return nil, err
	}

	respMap, err := jsonquery.Convertor.ObjectToMap(dataNode)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     respMap,
	}, nil
}
