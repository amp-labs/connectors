package gitlab

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
	perPageQuery      = "per_page"
	pageQuery         = "page"
	ownedQuery        = "owned"
	membershipQuery   = "membership"
	updatedAfterQuery = "updated_after"
	metadataPageSize  = "1"
	pageSize          = "100"
	projects          = "projects"
)

func (c *Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(perPageQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleHandlerResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	var firstRecord map[string]any

	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	switch objectResponders.Has(objectName) {
	case true:
		data, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		if len(*data) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		firstRecord = *data
	default:
		// In this case the response is an array, we unmarshal and assign the firstRecord
		// to our firstRecord variable tracker.
		data, err := common.UnmarshalJSON[[]map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		if len(*data) == 0 {
			return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
		}

		firstRecord = (*data)[0]
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url.WithQueryParam(pageQuery, params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) constructReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.ObjectName == projects {
		// sets owned=true&membership=true
		url.WithQueryParam(ownedQuery, "true")
		url.WithQueryParam(membershipQuery, "true")
	}

	if !params.Since.IsZero() {
		url.WithQueryParam(updatedAfterQuery, params.Since.Format(time.RFC3339))

		if params.ObjectName == projects {
			// This is required for reading Projects, if sice is provided.
			// ref: https://docs.gitlab.com/api/projects/#list-projects
			url.WithQueryParam("order_by", "updated_at")
		}
	}

	url.WithQueryParam(perPageQuery, pageSize)

	return url, nil
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
		nextRecordsURL(response.Headers.Get("X-Next-Page")),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
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

	resp, err := jsonquery.New(body).ObjectRequired("")
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
