package calendly

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/* The Calendly API provides event time filtering but not created/updated time filtering for most resources.
   To enable incremental reads, we can implement client-side filtering after full retrieval,
   comparing records against the 'since' timestamp. Currently not supported yet.
*/

const (
	organization = "organization"
	user         = "user"
	updatedAt    = "updated_at"
	countQuery   = "count"
	PageSize     = "100"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (string, error) { // nolint: cyclop
	var (
		url string
		err error
	)

	urlBuilder, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return "", err
	}

	// Retrieve 100 records per API call.
	urlBuilder.WithQueryParam(countQuery, PageSize)

	if requiresOrgURIQueryParam.Has(params.ObjectName) {
		urlBuilder.WithQueryParam(organization, c.orgURI)
	}

	if requiresUserURIQueryParam.Has(params.ObjectName) {
		urlBuilder.WithQueryParam(user, c.userURI)
	}

	if (!params.Since.IsZero()) && EndpointWithUpdatedAtParam.Has(params.ObjectName) {
		urlBuilder.WithQueryParam("sort", "updated_at:asc")
	}

	url = urlBuilder.String()

	if params.NextPage != "" {
		url = params.NextPage.String()
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	node, hasData := response.Body()
	if !hasData {
		// this should never occur as the API responds with a contents
		// having body regardless of status code.
		return &common.ReadResult{
			Rows: 0,
			Data: []common.ReadResultRow{},
			Done: true,
		}, nil
	}

	if !params.Since.IsZero() {
		if EndpointWithUpdatedAtParam.Has(params.ObjectName) {
			return manualIncrementalSync(node, dataKey, params, updatedAt, time.RFC3339, nextRecordsURL)
		}
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(dataKey),
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

// Manual incremental synchronization implementation for Calendly
//
// Calendly lacks native incremental sync support. This function iterates through records
// and returns those created or updated after the specified timestamp.
func manualIncrementalSync(node *ajson.Node, recordsKey string, config common.ReadParams, //nolint:cyclop
	timestampKey string, timestampFormat string, nextPageFunc common.NextPageFunc,
) (*common.ReadResult, error) {
	records, nextPage, err := readhelper.FilterSortedRecords(node, recordsKey,
		config.Since, timestampKey, timestampFormat, nextPageFunc)
	if err != nil {
		return nil, err
	}

	rows, err := common.GetMarshaledData(records, config.Fields.List())
	if err != nil {
		return nil, err
	}

	var done bool
	if nextPage == "" {
		done = true
	}

	return &common.ReadResult{
		Rows:     int64(len(records)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     done,
	}, nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		payload = params.RecordData
		method  = http.MethodPost
	)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(payload)
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

	resp, err := jsonquery.New(body).ObjectRequired("resource")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Data:    data,
	}, nil
}
