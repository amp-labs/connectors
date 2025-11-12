package highlevelwhitelabel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	if objectsWithLocationIdInParam.Has(objectName) {
		url.WithQueryParam("locationId", c.locationId)
	}

	if objectWithAltTypeAndIdQueryParam.Has(objectName) {
		url.WithQueryParam("altId", c.locationId)
		url.WithQueryParam("altType", "location")
	}

	if paginationObjects.Has(objectName) {
		url.WithQueryParam("limit", "1")

		if objectWithSkipQueryParam.Has(objectName) {
			url.WithQueryParam("skip", "0")
		} else {
			url.WithQueryParam("offset", "0")
		}
	}

	// For single-segment paths (e.g., "businesses"), the URL must have a trailing slash at the end.
	// Example: https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location
	//
	// For multi-segment paths (e.g., "calendars/groups"), the URL does not require a trailing slash.
	// Example: https://highlevel.stoplight.io/docs/integrations/89e47b6c05e67-get-groups
	if !(strings.Contains(objectName, "/")) {
		urlRaw, err := url.ToURL()
		if err != nil {
			return nil, err
		}

		urlRaw.Path = urlRaw.Path + "/" // nolint:gocritic

		url, err = urlbuilder.FromRawURL(urlRaw)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Version", apiVersion)

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	node, ok := response.Body() // nolint:varnamelen
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	nodePath := objectsNodePath.Get(objectName)

	objectResponse, err := jsonquery.New(node).ArrayRequired(nodePath)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ArrayToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	for field := range data[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    "other",
			ProviderType: "",
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest( // nolint:cyclop
	ctx context.Context,
	params common.ReadParams,
) (*http.Request, error) {
	var (
		nextPage int
		err      error
	)

	if params.NextPage != "" {
		// Parse the page number from NextPage
		nextPage, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if objectsWithLocationIdInParam.Has(params.ObjectName) {
		url.WithQueryParam("locationId", c.locationId)
	}

	if objectWithAltTypeAndIdQueryParam.Has(params.ObjectName) {
		url.WithQueryParam("altId", c.locationId)
		url.WithQueryParam("altType", "location")
	}

	if paginationObjects.Has(params.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(defaultPageSize))

		if objectWithSkipQueryParam.Has(params.ObjectName) {
			url.WithQueryParam("skip", strconv.Itoa(nextPage))
		} else {
			url.WithQueryParam("offset", strconv.Itoa(nextPage))
		}
	}

	// For single-segment paths (e.g., "businesses"), the URL must have a trailing slash at the end.
	// Example: https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location
	//
	// For multi-segment paths (e.g., "calendars/groups"), the URL does not require a trailing slash.
	// Example: https://highlevel.stoplight.io/docs/integrations/89e47b6c05e67-get-groups
	if !(strings.Contains(params.ObjectName, "/")) {
		urlRaw, err := url.ToURL()
		if err != nil {
			return nil, err
		}

		urlRaw.Path = urlRaw.Path + "/" // nolint:gocritic

		url, err = urlbuilder.FromRawURL(urlRaw)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Version", apiVersion)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	var (
		offset int
		err    error
	)

	if params.NextPage.String() != "" {
		offset, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(objectsNodePath.Get(params.ObjectName)),
		makeNextRecord(offset, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	if !(strings.Contains(params.ObjectName, "/")) {
		urlRaw, err := url.ToURL()
		if err != nil {
			return nil, err
		}

		urlRaw.Path = urlRaw.Path + "/" // nolint:gocritic

		url, err = urlbuilder.FromRawURL(urlRaw)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Version", apiVersion)

	return req, nil
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

	record, err := jsonquery.New(body).ObjectOptional(writeObjectsNodePath.Get(params.ObjectName))
	if err != nil {
		return nil, err
	}

	recordId := ""

	if writeObjectsWithIdField.Has(params.ObjectName) {
		recordId = "id"
	}

	if writeObjectsWithUnderscoreIdField.Has(params.ObjectName) {
		recordId = "_id"
	}

	recordID, err := jsonquery.New(record).StrWithDefault(recordId, "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(record)
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
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	var bodyParam []byte

	// Some endpoints requires locationId in the query param.
	// refer https://highlevel.stoplight.io/docs/integrations/96bc73da716e8-delete-relation.
	if deleteObjectWithLocationIdQueryParam.Has(params.ObjectName) {
		url.WithQueryParam("locationId", c.locationId)
	}

	// Some endpoints requires altId and altType in the query param.
	// refer https://highlevel.stoplight.io/docs/integrations/af9fb9b428e74-delete-invoice.
	if objectWithAltTypeAndIdQueryParam.Has(params.ObjectName) {
		url.WithQueryParam("altId", c.locationId)
		url.WithQueryParam("altType", "location")
	}

	// Some objects requires altId and altType in the body param.
	// refer https://highlevel.stoplight.io/docs/integrations/3c4f7a7d1d4d9-delete-estimate-template.
	if objectWithAltTypeAndIdBodyParam.Has(params.ObjectName) {
		param := map[string]string{
			"altId":   c.locationId,
			"altType": "location",
		}

		bodyParam, err = json.Marshal(param)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), bytes.NewReader(bodyParam))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Version", apiVersion)

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

	return &common.DeleteResult{
		Success: true,
	}, nil
}
