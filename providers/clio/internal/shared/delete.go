package shared //nolint:revive

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// BuildDeleteParams carries dependencies for constructing a Clio delete HTTP request.
type BuildDeleteParams struct {
	BaseURL     string
	APIPath     string
	Module      common.ModuleID
	Params      common.DeleteParams
	FindURLPath func(common.ModuleID, string) (string, error)
}

// BuildDeleteRequest builds DELETE for a single record (Manage `.json` paths vs Grow paths).
//
// Docs: https://docs.developers.clio.com/api-docs/
func BuildDeleteRequest(ctx context.Context, buildParams BuildDeleteParams) (*http.Request, error) {
	params := buildParams.Params
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	path, err := buildParams.FindURLPath(buildParams.Module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	rel := strings.TrimPrefix(path, "/")

	var requestURL *urlbuilder.URL

	switch {
	case strings.HasSuffix(rel, ".json"):
		basePath := strings.TrimSuffix(rel, ".json")
		requestURL, err = urlbuilder.New(buildParams.BaseURL, buildParams.APIPath, basePath+"/"+params.RecordId+".json")
	default:
		requestURL, err = urlbuilder.New(buildParams.BaseURL, buildParams.APIPath, rel, params.RecordId)
	}

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	for _, h := range params.Headers {
		req.Header.Add(h.Key, h.Value)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// ParseDeleteResponse maps a successful Clio delete HTTP response into common.DeleteResult.
// Clio returns 204 No Content (empty body) for successful DELETE; common.ParseJSONResponse handles that.
func ParseDeleteResponse(_ common.DeleteParams, resp *common.JSONHTTPResponse) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(resp.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
