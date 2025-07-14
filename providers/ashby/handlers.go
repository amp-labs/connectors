package ashby

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "limit"
	pageSize    = "100"
	pageKey     = "cursor"
	sinceKey    = "createdAfter"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	body := buildRequestbody(params)

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func buildRequestbody(params common.ReadParams) map[string]any {
	body := make(map[string]any)

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		body[sinceKey] = params.Since.UnixMilli()
	}

	if supportPagination.Has(params.ObjectName) && params.NextPage != "" {
		body[pageKey] = params.NextPage
		body[pageSizeKey] = pageSize
	}

	return body
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		makeNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

// Note: Ashby API uses POST method for all operations (create/update/delete).
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	objectName := params.ObjectName

	if len(params.RecordId) > 0 {
		// Add .update as a suffix if it’s an update operation.
		objectName += ".update"
	} else {
		// Add .create as a suffix if it’s a create operation.
		objectName += ".create"
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
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
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	success, err := jsonquery.New(node).BoolWithDefault("success", true)
	if err != nil {
		return nil, err
	}

	if !success {
		return handleErrorResponse(node)
	}

	results, err := jsonquery.New(node).ObjectRequired("results")
	if err != nil {
		//nolint:nilerr
		return &common.WriteResult{Success: true}, nil
	}

	rawID, err := jsonquery.New(node, "results").StringOptional("id")
	if err != nil {
		//nolint:nilerr
		return &common.WriteResult{Success: true}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(results)
	if err != nil {
		//nolint:nilerr
		return &common.WriteResult{
			Success:  true,
			RecordId: *rawID,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *rawID,
		Data:     data,
	}, nil
}

func handleErrorResponse(node *ajson.Node) (*common.WriteResult, error) {
	errors, err := jsonquery.New(node).ArrayOptional("errors")
	if err != nil {
		return &common.WriteResult{Success: false}, nil //nolint:nilerr
	}

	errorArr, err := jsonquery.Convertor.ArrayToObjects(errors)
	if err != nil {
		//nolint:nilerr
		return &common.WriteResult{Success: false}, nil
	}

	return &common.WriteResult{
		Success: false,
		Errors:  errorArr,
	}, nil
}
