package shared //nolint:revive

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// BuildWriteParams carries dependencies for constructing a Clio write HTTP request.
type BuildWriteParams struct {
	BaseURL     string
	APIPath     string
	Module      common.ModuleID
	Params      common.WriteParams
	FindURLPath func(common.ModuleID, string) (string, error)
}

// writeURL is Manage …/things.json vs …/things/{id}.json; Grow appends recordId as a path segment.
func writeURL(baseURL, apiPath, rel, recordID string) (*urlbuilder.URL, error) {
	if recordID == "" {
		return urlbuilder.New(baseURL, apiPath, rel)
	}

	if before, ok := strings.CutSuffix(rel, ".json"); ok {
		base := before

		return urlbuilder.New(baseURL, apiPath, base+"/"+recordID+".json")
	}

	return urlbuilder.New(baseURL, apiPath, rel, recordID)
}

// BuildWriteRequest builds POST (create) or PATCH (update) for Clio Manage / Grow JSON APIs.
//
// Docs: https://docs.developers.clio.com/api-docs/
func BuildWriteRequest(ctx context.Context, buildParams BuildWriteParams) (*http.Request, error) {
	path, err := buildParams.FindURLPath(buildParams.Module, buildParams.Params.ObjectName)
	if err != nil {
		return nil, err
	}

	rel := strings.TrimLeft(path, "/")

	record, err := buildParams.Params.GetRecord()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{"data": record})
	if err != nil {
		return nil, err
	}

	u, err := writeURL(buildParams.BaseURL, buildParams.APIPath, rel, buildParams.Params.RecordId)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if buildParams.Params.RecordId != "" {
		method = http.MethodPatch
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ParseWriteResponse maps a Clio JSON write response into common.WriteResult (saved record in "data", per API docs).
func ParseWriteResponse(params common.WriteParams, resp *common.JSONHTTPResponse) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	data, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	recordMap, err := jsonquery.Convertor.ObjectToMap(data)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(data).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     recordMap,
	}, nil
}
