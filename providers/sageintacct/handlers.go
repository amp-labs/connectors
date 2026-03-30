package sageintacct

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
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

// nolint:lll
// We flatten the fields, groups, and refs into a single map of field metadata
// with JSONPath bracket notation keys for nested fields.
// We require to mention the fields explicitly in the read requests.
// Sage doesn't support wildcard selection of fields or nested field selection.
// So we need to ensure all fields are explicitly specified in the request like $['audit']['createdByUser']['key'].
// Due to which we are flattening the metadata here.
//
// IMPORTANT: Lists (array fields like "locations", "departments") are NOT queryable in SageIntacct.
// According to SageIntacct API documentation:
// "A field may not be queryable if it returns a list of data, which is not supported for querying"
// Ref: https://developer.sage.com/intacct/docs/1/sage-intacct-rest-api/api-essentials/query-service#troubleshooting-and-faqs
//
// Example:
/*
metadata response structure:
   "fields": {
     "id": { ... },
     "name": { ... }
   },
   "groups": {
     "audit": {
       "fields": {
         "createdByUser": {
           "fields": { "key": {...}, "name": {...} }
         },
         "createdDate": { ... }
       }
     }
   },
   "refs": {
     "contact": {
       "fields": { "id": {...}, "firstName": {...} }
     }
   },
   "lists": {
     "locations": {
       "fields": { "id": {...}, "key": {...} }  // NOT queryable - will be ignored
     }
   }

   Flattened output:
   "$['id']": { ... },
   "$['name']": { ... },
   "$['audit']['createdByUser']['key']": { ... },
   "$['audit']['createdByUser']['name']": { ... },
   "$['audit']['createdDate']": { ... },
   "$['contact']['id']": { ... },
   "$['contact']['firstName']": { ... }
   // Note: "locations" is NOT included because lists are not queryable
*/

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

	// Flatten top-level fields
	topLevelFields := flattenFields(nil, res.Result.Fields)
	maps.Copy(objectMetadata.Fields, topLevelFields)

	// Flatten groups (nested objects like audit, etc.)
	if len(res.Result.Groups) > 0 {
		groupFields := flattenGroups(nil, res.Result.Groups)
		maps.Copy(objectMetadata.Fields, groupFields)
	}

	// Flatten refs (reference objects)
	if len(res.Result.Refs) > 0 {
		refFields := flattenRefs(nil, res.Result.Refs)
		maps.Copy(objectMetadata.Fields, refFields)
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
	dotNotationFieldNames, err := convertFieldsToDotNotation(params.Fields.List())
	if err != nil {
		return nil, err
	}

	var fieldNameSet datautils.Set[string]
	if len(dotNotationFieldNames) > 0 {
		fieldNameSet = datautils.NewStringSet(dotNotationFieldNames...)
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("ia::result"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		fieldNameSet,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiPrefix, apiVersion, "objects", params.ObjectName)
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

	// example: https://developer.sage.com/intacct/docs/openapi/gl/general-ledger.budget/tag/Budgets
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
