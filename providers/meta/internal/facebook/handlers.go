package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const apiVersion = "v19.0"

type ResponseData struct {
	Data []map[string]any `json:"data"`
}

func (c *Adapter) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := c.constructURL(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// Adding accept because this connector sending response text/javascript.
	request.Header.Add("Accept", "*/*")

	return request, nil
}

func (c *Adapter) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[ResponseData](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(resp.Data) == 0 {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	for field := range resp.Data[0] {
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

func (c *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	urlPath := c.constructURL(params.ObjectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(defaultPageSize))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "*/*")

	return request, nil
}

func (c *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	urlPath := c.constructURL(params.ObjectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "*/*")

	return request, nil
}

func (c *Adapter) parseWriteResponse(
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
