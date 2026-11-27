package freshchat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiVersion = "v2"
	pageSize   = "100"
)

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	path := params.ObjectName
	if readPath, found := readPaths[params.ObjectName]; found {
		path = readPath
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	if singleRecordObjects.Has(params.ObjectName) {
		// There is nothing to page through or to filter.
		return url, nil
	}

	addReadQueryParams(url, params)

	return url, nil
}

// addReadQueryParams sets the page size and the time range of a list of records.
func addReadQueryParams(url *urlbuilder.URL, params common.ReadParams) {
	if params.PageSize != 0 {
		url.WithQueryParam("items_per_page", strconv.Itoa(params.PageSize))
	} else {
		url.WithQueryParam("items_per_page", pageSize)
	}

	// users requires 1 filtering query parameter.
	isDefaultTimeRange := params.Since.IsZero() && params.Until.IsZero()
	if params.ObjectName == objectUsers && isDefaultTimeRange {
		url.WithQueryParam("created_from", time.Unix(0, 0).UTC().Format(time.RFC3339))
	}

	if !params.Since.IsZero() {
		if supportFilteringByTime.Has(params.ObjectName) {
			url.WithQueryParam("updated_from", params.Since.Format(time.RFC3339))
		}
	}

	if !params.Until.IsZero() {
		if supportFilteringByTime.Has(params.ObjectName) {
			url.WithQueryParam("updated_to", params.Until.Format(time.RFC3339))
		}
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	urlbuild, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, urlbuild.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(c.ProviderInfo().BaseURL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut

		if isAgentStatusUpdate(params) {
			method = http.MethodPatch
		}
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// isAgentStatusUpdate reports whether the write changes nothing but the status of an agent.
// Freshchat has a dedicated endpoint for it, PATCH /v2/agents/{agent_id}, while every other
// property of an agent is updated with PUT on the same path.
// https://developers.freshchat.com/api/#update_agent_status
func isAgentStatusUpdate(params common.WriteParams) bool {
	if params.ObjectName != objectAgents || params.RecordId == "" {
		return false
	}

	record, isMap := params.RecordData.(map[string]any)
	if !isMap || len(record) != 1 {
		return false
	}

	_, hasStatus := record[agentStatusField]

	return hasStatus
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// An update answers with no body, the record being the one that was addressed.
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	resp, err := jsonquery.New(body).ObjectRequired("")
	if err != nil {
		return nil, err
	}

	idField := defaultIDField
	if field, found := writeIDFields[params.ObjectName]; found {
		idField = field
	}

	recordId, err := jsonquery.New(resp).StringOptional(idField)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	// Not every response names the record, in which case it is the one that was addressed.
	id := params.RecordId
	if recordId != nil {
		id = *recordId
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: id,
		Data:     data,
	}, nil
}
