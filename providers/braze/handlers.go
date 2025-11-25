package braze

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type catalogPayload struct {
	Catalogs []map[string]any `json:"catalogs"`
}

var ErrInvalidData = errors.New("invalid request data provided")

func (c *Connector) metadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	path := objectName

	// maps objectName to the braze APIs endpoints path.
	if objectName, exists := readEndpointsByObject[objectName]; exists {
		path = objectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	// sets single page request for metadata response.
	url.WithQueryParam(limitQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	path := objectName

	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, err
	}

	if endpoint, exists := readEndpointsByObject[objectName]; exists {
		path = endpoint
	}

	fld := dataFields.Get(path)

	rcds, okay := (*data)[fld].([]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(rcds) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	// Iterate over the first record.
	firstRecord, ok := rcds[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for fld := range firstRecord {
		// TODO fix deprecated
		objectMetadata.FieldsMap[fld] = fld // nolint:staticcheck
	}

	return &objectMetadata, nil
}

func (c *Connector) constructReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	path := params.ObjectName

	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	if obj, exists := readEndpointsByObject[params.ObjectName]; exists {
		path = obj
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if offsetPaginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam(offset, "1")
	}

	return url, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	if err := setSinceQuery(params, url); err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	path := params.ObjectName

	// map objectName to the braze APIs endpoints.
	if obj, exists := readEndpointsByObject[params.ObjectName]; exists {
		// map the objectName to the appropriate endpoint
		path = obj
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	prevPage := request.URL.Query().Get(page)

	// map objectName to the braze APIs endpoints.
	if objectName, exists := readEndpointsByObject[params.ObjectName]; exists {
		// map the objectName to the appropriate endpoint
		params.ObjectName = objectName
	}

	return common.ParseResult(response,
		common.ExtractRecordsFromPath(dataFields.Get(params.ObjectName)),
		getNextRecordsURL(params.ObjectName, prevPage, url, response),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) constructCatalogPayload(recordData any) ([]byte, error) {
	payload := catalogPayload{
		Catalogs: make([]map[string]any, 1),
	}

	data, ok := recordData.(map[string]any)
	if !ok {
		return nil, ErrInvalidData
	}

	payload.Catalogs[0] = data

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return jsonData, nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		jsonData []byte
		err      error
		url      *urlbuilder.URL
		method   = http.MethodPost
	)

	// map objectName to the braze APIs endpoints.
	path, err := constructWritePath(params)
	if err != nil {
		return nil, err
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	if params.ObjectName == "catalogs" {
		jsonData, err = c.constructCatalogPayload(params.RecordData)
		if err != nil {
			return nil, err
		}
	} else {
		jsonData, err = json.Marshal(params.RecordData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal record data: %w", err)
		}
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	return &common.WriteResult{
		Success: true,
	}, nil
}
