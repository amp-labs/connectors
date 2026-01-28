package pardot

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
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

	url.WithQueryParam("fields", strings.Join(params.Fields.List(), ","))
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

// incrementalSinceQuery is a registry of object name to the query parameter used for performing incremental reading.
var incrementalSinceQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtAfterOrEqualTo",
}

// incrementalUntilQuery is a registry of object name to the query parameter used for performing incremental reading.
var incrementalUntilQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtBeforeOrEqualTo",
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("values"),
		func(node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("nextPageUrl", "")
		},
		common.GetMarshaledData,
		params.Fields,
	)
}
