package campaignmonitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const APIVersion = "v3.3"

type ResponseData struct {
	Results []map[string]any `json:"Results,omitempty"` // nolint:tagliatelle
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, objectName+".json")
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

	// Direct Response
	resp, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, err
	}

	if len(*resp) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	record := *resp

	for field := range record[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	// Here is the link explaining why we add .json after the objectName: https://www.campaignmonitor.com/api/v3-3/clients/
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, params.ObjectName+".json")
	if err != nil {
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
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(""),
		getNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

func getNextRecordsURL(_ *ajson.Node) (string, error) {
	// Pagination is not supported for this provider.
	return "", nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, params.ObjectName+".json")
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
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

	value, err := body.Value()
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		// This occurs when the API returns a raw string as the response body,
		// typically just the id of the created/updated object.
		// For example: "42f13de0-021c-11ef-b57f-0242ac120003"
		return &common.WriteResult{
			Success:  true,
			RecordId: v,
			Errors:   nil,
		}, nil

	default:
		// This occurs when the API responds with a full JSON object containing additional metadata or nested structures.
		// For example: {
		//  "EmailAddress": "sally@sparrow.com"
		// }
		data, err := jsonquery.Convertor.ObjectToMap(body)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{
			Success: true,
			Errors:  nil,
			Data:    data,
		}, nil
	}
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, params.ObjectName, params.RecordId+".json")
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
