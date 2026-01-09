package g2

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
)

const (
	restAPIVersion = "api/v2"
)

type Response struct {
	Data  []record       `json:"data"`
	Links map[string]any `json:"links"`
}

type record struct {
	Id            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    map[string]any `json:"attributes"`
	Relationships map[string]any `json:"relationships"`
}

type data struct {
	Type       string         `json:"string"`
	Id         string         `json:"id,omitempty"`
	Attributes map[string]any `json:"attributes"`
}

type writeRequest struct {
	Data data `json:"data"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	path, err := PathsConfig(c.productId, objectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)
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
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[Response](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	// Add attributes fields to metadata
	firstRecord := resp.Data[0].Attributes
	for fld, val := range firstRecord {
		objectMetadata.Fields[fld] = common.FieldMetadata{
			DisplayName: fld,
			ValueType:   inferValueTypeFromData(val),
		}
	}

	// Add the id of the data layer into fields.
	objectMetadata.Fields["id"] = common.FieldMetadata{
		DisplayName: "Id",
		ValueType:   common.ValueTypeString,
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
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
		records(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		jsonData []byte
		method   = http.MethodPost
	)

	path, err := writePath(c.productId, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	if params.ObjectName == PathProductMappings { //nolint: nestif
		var req writeRequest

		if params.RecordId != "" {
			req.Data.Id = params.RecordId
		}

		reqData, ok := params.RecordData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected the request body to be an object, writing to %s", params.ObjectName) //nolint:err113
		}

		req.Data.Type = PathProductMappings
		req.Data.Attributes = reqData

		jsonData, err = json.Marshal(params.RecordData)
		if err != nil {
			return nil, err
		}
	} else {
		jsonData, err = json.Marshal(params.RecordData)
		if err != nil {
			return nil, err
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
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	resp, err := jsonquery.New(body).ObjectRequired("data")
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
