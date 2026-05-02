package shared //nolint:revive

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const MaxPageSize = "200" // Docs: https://docs.developers.clio.com/api-docs/clio-manage/paging

type BuildReadParams struct {
	BaseURL     string
	APIPath     string
	Module      common.ModuleID
	Params      common.ReadParams
	FindURLPath func(common.ModuleID, string) (string, error)
	// ObjectsNoUpdatedSince lists objects that do NOT support the provider-side `updated_since` query param.
	ObjectsNoUpdatedSince datautils.Set[string]
}

func BuildReadRequest(ctx context.Context, buildParams BuildReadParams) (*http.Request, error) {
	params := buildParams.Params
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := buildParams.FindURLPath(buildParams.Module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(buildParams.BaseURL, buildParams.APIPath, path)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, MaxPageSize)

	url.WithQueryParam("limit", pageSize)

	// Cursor mode requires order=id(asc), Docs: https://docs.developers.clio.com/api-docs/clio-manage/paging
	url.WithQueryParam("order", "id(asc)")

	if buildParams.ObjectsNoUpdatedSince.Has(params.ObjectName) && !params.Fields.Has("updated_at") {
		params.Fields.AddOne("updated_at")
	}
	// fields selects which record fields to return if omitted only defaults like id and etag are returned
	// Docs: https://docs.developers.clio.com/api-docs/clio-manage/fields
	if fld := strings.Join(params.Fields.List(), ","); fld != "" {
		url.WithUnencodedQueryParam("fields", fld)
	}

	if !params.Since.IsZero() && !buildParams.ObjectsNoUpdatedSince.Has(params.ObjectName) {
		url.WithQueryParam("updated_since", params.Since.UTC().Format(time.RFC3339))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func ParseReadResponse(
	params common.ReadParams,
	resp *common.JSONHTTPResponse,
	module common.ModuleID,
	lookupArrayFieldName func(common.ModuleID, string) string,
	objectsNoUpdatedSince datautils.Set[string],
) (*common.ReadResult, error) {
	responseKey := lookupArrayFieldName(module, params.ObjectName)
	if responseKey == "" {
		responseKey = params.ObjectName
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(responseKey),
		makeFilterFunc(params, objectsNoUpdatedSince),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func makeFilterFunc(
	params common.ReadParams,
	objectsNoUpdatedSince datautils.Set[string],
) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL()

	if !objectsNoUpdatedSince.Has(params.ObjectName) {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		"updated_at",
		time.RFC3339,
		nextPageFunc,
	)
}

// Example response:
//
//	{
//		"data": { ... },
//		"meta": {
//		  "paging": {
//			"previous": "https://app.clio.com/api/v4/contacts?fields=...",
//			"next": "https://app.clio.com/api/v4/contacts?fields=..."
//		  }
//		},
//	  }
func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node, "meta", "paging").StrWithDefault("next", "")
	}
}
