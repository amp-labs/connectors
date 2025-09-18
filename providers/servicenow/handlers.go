package servicenow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
)

type responseData struct {
	Result []map[string]any `json:"result"`
	// Other fields
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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

	res, err := common.UnmarshalJSON[responseData](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(res.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// Using the first result data to generate the metadata.
	for k := range res.Result[0] {
		// TODO fix deprecated
		objectMetadata.FieldsMap[k] = k // nolint:staticcheck
	}

	return &objectMetadata, nil
}

func (c *Connector) constructReadURL(params common.ReadParams) (string, error) {
	if params.NextPage != "" {
		return params.NextPage.String(), nil
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(response,
		common.ExtractRecordsFromPath("result"),
		getNextRecordsURL(response),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	logging.With(ctx, "connector", providers.ServiceNow)

	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
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

	result, err := jsonquery.New(body).ObjectRequired("result")
	if err != nil {
		logging.Logger(ctx).Error("failed to parse write response", "object", params.ObjectName, "body", body, "err", err.Error()) //nolint:lll

		return &common.WriteResult{Success: true}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(result)
	if err != nil {
		logging.Logger(ctx).Error("failed to convert result object to map", "object", params.ObjectName, "err", err.Error())

		return &common.WriteResult{Success: true}, nil
	}

	return &common.WriteResult{
		Success: true,
		Errors:  nil,
		Data:    data,
	}, nil
}

func getNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}
