package loxo

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.AgencySlug, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

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

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	nodePath := objectsNodePath.Get(objectName)

	res, err := jsonquery.New(body).ArrayOptional(nodePath)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return nil, common.ErrMissingMetadata
	}

	for field := range record[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.AgencySlug, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if paginationObjects.Has(params.ObjectName) {
		url.WithQueryParam("per_page", strconv.Itoa(defaultPageSize))
	}

	if incrementalReadobjects.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("created_at_start", params.Since.Format(time.DateOnly))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("created_at_end", params.Until.Format(time.DateOnly))
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(objectsNodePath.Get(params.ObjectName)),
		makeNextRecordsURL(url, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}
