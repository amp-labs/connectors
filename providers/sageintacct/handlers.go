package sageintacct

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
)

const (
	apiPrefix       = "ia/api"
	apiVersion      = "v1"
	defaultPageSize = 500
	pageSizeParam   = "size"
	pageParam       = "start"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiPrefix, apiVersion, "services/core/model")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("name", objectName)
	url.WithQueryParam("version", apiVersion)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetter(objectName),
	}

	bodyNode, ok := response.Body()
	if !ok {
		return nil, common.ErrFailedToUnmarshalBody
	}

	resultNode, err := bodyNode.GetKey("ia::result")
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	// api returns array when object is not supported.
	if resultNode.IsArray() {
		return nil, common.ErrObjectNotSupported
	}

	res, err := common.UnmarshalJSON[SageIntacctMetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	for fieldName, fieldDef := range res.Result.Fields {
		objectMetadata.Fields[fieldName] = common.FieldMetadata{
			DisplayName:  naming.CapitalizeFirstLetterEveryWord(fieldName),
			ValueType:    mapSageIntacctTypeToValueType(fieldDef.Type),
			ProviderType: fieldDef.Type,
			ReadOnly:     fieldDef.ReadOnly,
			Values:       mapValuesFromEnum(fieldDef),
		}
	}

	if len(res.Result.Groups) > 0 {
		for groupName := range res.Result.Groups {
			objectMetadata.Fields[groupName] = common.FieldMetadata{
				DisplayName:  naming.CapitalizeFirstLetterEveryWord(groupName),
				ValueType:    common.ValueTypeOther,
				ProviderType: "object",
				ReadOnly:     false,
				Values:       []common.FieldValue{},
			}
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiPrefix, apiVersion, "services/core/query")
	if err != nil {
		return nil, err
	}

	body, err := buildReadBody(params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("ia::result"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
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

	resp, err := jsonquery.New(body).ObjectOptional("ia::result")
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(resp).StrWithDefault("key", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		Data:     data,
		Errors:   nil,
		RecordId: recordID,
	}, nil
}
