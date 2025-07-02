package campaignmonitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const APIVersion = "v3.3"

type ResponseData struct {
	Results []map[string]any `json:"Results,omitempty"` // nolint:tagliatelle
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
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

	switch objectName {
	// Below two objects having the response which is embedded with the "Results" key.
	case "suppressionlist", "campaigns":
		resp, err := common.UnmarshalJSON[ResponseData](response)
		if err != nil {
			return nil, err
		}

		if len(resp.Results) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		for field := range resp.Results[0] {
			objectMetadata.FieldsMap[field] = field
		}
	default:
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
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", strconv.Itoa(defaultPageSize))

	// Only campaigns objects supports pagination with query param sentFromDate and sentToDate.
	if params.ObjectName == "campaigns" {
		if !params.Since.IsZero() {
			url.WithQueryParam("sentFromDate", params.Since.Format(time.DateOnly))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("sentToDate", params.Until.Format(time.DateOnly))
		}
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

	if endpointsWtihResultsPath.Has(params.ObjectName) {
		nodePath = "Results"
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.constructWriteURL(params.ObjectName)
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
		// This occurs when the API returns a raw string as the response body, typically just the id of the created/updated object.
		// For example: "42f13de0-021c-11ef-b57f-0242ac120003"
		return &common.WriteResult{
			Success:  true,
			RecordId: v,
			Errors:   nil,
		}, nil

	default:
		// This occurs when the API responds with a full JSON object containing additional metadata or nested structures.
		// For example: {
		//   "campaign": {
		//     "id": "42f13de0-021c-11ef-b57f-0242ac120003",
		//     "name": "New Campaign"
		//   }
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
