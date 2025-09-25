package chorus

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

const apiVersion = "v1"

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

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

	res, err := jsonquery.New(body).ArrayRequired("data")
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// helper to create FieldMetadata
	newField := func(name string) common.FieldMetadata {
		return common.FieldMetadata{
			DisplayName:  name,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			ReadOnly:     false,
			Values:       nil,
		}
	}

	for field, value := range record[0] {
		if field == "attributes" {
			if subfields, ok := value.(map[string]any); ok {
				for subfield := range subfields {
					objectMetadata.Fields[subfield] = newField(subfield)
				}
			} else {
				return nil, common.ErrMissingFields
			}

			continue
		}

		objectMetadata.Fields[field] = newField(field)
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if PaginationObject.Has(params.ObjectName) {
		url.WithQueryParam("page[size]", strconv.Itoa(PageSize))

		if params.NextPage != "" {
			url.WithQueryParam("page[number]", params.NextPage.String())
		}
	}

	if IncrementalObjectQueryParam.Has(params.ObjectName) {
		startDate := params.Since.Format(time.RFC3339)

		endDate := params.Until.Format(time.RFC3339)

		url.WithQueryParam(IncrementalObjectQueryParam.Get(params.ObjectName), startDate+":"+endDate)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	var (
		nextPage int
		err      error
	)

	if params.NextPage != "" {
		nextPage, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		makeNextRecord(nextPage),
		DataMarshall(response),
		params.Fields,
	)
}
