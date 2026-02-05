package pardot

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// DefaultPageSize is a relative limit. In fact the largest page size allowed is 100,000.
// Reference:
// https://developer.salesforce.com/docs/marketing/pardot/guide/version5overview.html#pagination
const DefaultPageSize = "1000"

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Pardot-Business-Unit-Id", a.businessUnitID)

	return req, nil
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	objectNameLower := strings.ToLower(params.ObjectName)

	url, err := a.getURL(objectNameLower)
	if err != nil {
		return nil, err
	}

	applyFieldsAndOrdering(url, params)
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, DefaultPageSize))

	if !params.Since.IsZero() {
		if query, ok := incrementalSinceQuery[objectNameLower]; ok {
			url.WithQueryParam(query, datautils.Time.FormatRFC3339WithOffset(params.Since))
		}
	}

	if !params.Until.IsZero() {
		if query, ok := incrementalUntilQuery[objectNameLower]; ok {
			url.WithQueryParam(query, datautils.Time.FormatRFC3339WithOffset(params.Until))
		}
	}

	return url, nil
}

// applyFieldsAndOrdering configures the request URL with the correct `fields`
// and (when required) `orderBy` query parameters.
//
// It starts from the fields explicitly requested by the caller. If the target
// object supports connector-side time filtering (for example, filtering by
// createdAt or updatedAt), this function:
//
//   - ensures the required timestamp field is included in the `fields` query
//   - applies an ascending `orderBy` so results are returned in chronological order
//
// This is necessary because connector-side filtering can only be applied if the
// relevant field is present in the raw response payload.
//
// A set is used to avoid duplicate fields, and a copy is created to prevent
// mutating the original params.Fields slice.
func applyFieldsAndOrdering(url *urlbuilder.URL, params common.ReadParams) {
	fields := params.Fields.List()

	timeField, found := objectsFilterParam[params.ObjectName]
	if found {
		// Ensure the required createdAt/updatedAt field is included
		// so connector-side filtering can be applied.
		list := datautils.NewSetFromList(fields)
		list.AddOne(timeField)
		fields = list.List()

		// Enforce chronological order for stable filtering
		url.WithQueryParam("orderBy", timeField+" ASC")
	}

	url.WithQueryParam("fields", strings.Join(fields, ","))
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(params, resp,
		common.MakeRecordsFunc("values"),
		makeFilterFunc(params),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}
