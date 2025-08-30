package breakcold

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
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodGet

	if getEndpointsPostMethod.Has(objectName) {
		method = http.MethodPost

		url = url.AddPath("list")
	}

	return http.NewRequestWithContext(ctx, method, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	nodepath := ""

	// The endpoint has data nodePath in the response.
	// https://developer.breakcold.com/v3/api-reference/reminders/list-reminders-with-filters-and-pagination.
	if objectName == objectNameRemindersList {
		nodepath = "data"
	}

	//  The endpoint has leads as the nodePath in the response.
	//  https://developer.breakcold.com/v3/api-reference/leads/list-leads-with-pagination-and-filters.
	if objectName == objectNameLeadsList {
		nodepath = "leads"
	}

	res, err := jsonquery.New(body).ArrayOptional(nodepath)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	for field := range record[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if getEndpointsPostMethod.Has(params.ObjectName) {
		url = url.AddPath("list")

		body, err := constructRequestBody(params)
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	nodePath := ""

	// The endpoint has data nodePath in the response.
	// https://developer.breakcold.com/v3/api-reference/reminders/list-reminders-with-filters-and-pagination.
	if params.ObjectName == objectNameRemindersList {
		nodePath = "data"
	}

	//  The endpoint has leads as the nodePath in the response.
	//  https://developer.breakcold.com/v3/api-reference/leads/list-leads-with-pagination-and-filters.
	if params.ObjectName == objectNameLeadsList {
		nodePath = "leads"
	}

	var (
		nextPage int
		err      error
	)

	if params.NextPage.String() != "" {
		nextPage, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecordsURL(nodePath, nextPage),
		common.GetMarshaledData,
		params.Fields,
	)
}

func constructRequestBody(config common.ReadParams) ([]byte, error) {
	page := 0

	if len(config.NextPage) != 0 {
		nextPage, err := strconv.Atoi(config.NextPage.String())
		if err != nil {
			return nil, err
		}

		page = nextPage
	}

	body := map[string]any{
		"pagination": map[string]int{
			"page":      page,
			"page_size": pageSize,
		},
	}

	if config.ObjectName == objectNameRemindersList {
		body["cursor"] = page
	}

	return json.Marshal(body)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		urlRaw, err := url.ToURL()
		if err != nil {
			return nil, err
		}

		// Some update endpoints having plural objectname.
		// Ref https://developer.breakcold.com/v3/api-reference/leads/update-a-lead.
		urlRaw.Path = naming.NewPluralString(urlRaw.Path).String()

		url, err = urlbuilder.FromRawURL(urlRaw)
		if err != nil {
			return nil, err
		}

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
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
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

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL,
		naming.NewPluralString(params.ObjectName).String(),
		params.RecordId,
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
