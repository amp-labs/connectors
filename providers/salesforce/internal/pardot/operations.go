package pardot

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// DefaultPageSize is a relative limit. In fact the largest page size allowed is 100,000.
// Reference:
// https://developer.salesforce.com/docs/marketing/pardot/guide/version5overview.html#pagination
const DefaultPageSize = "1000"

func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	rsp, err := a.Client.Get(ctx, url.String(), common.Header{
		Key:   "Pardot-Business-Unit-Id",
		Value: a.BusinessUnitID,
	})
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.ExtractOptionalRecordsFromPath("values"),
		func(node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("nextPageUrl", "")
		},
		common.GetMarshaledData,
		params.Fields,
	)
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
	url.WithQueryParam("limit", DefaultPageSize)

	if !params.Since.IsZero() {
		if query, ok := incrementalQuery[objectNameLower]; ok {
			url.WithQueryParam(query, datautils.Time.FormatRFC3339WithOffset(params.Since))
		}
	}

	return url, nil
}

// incrementalQuery is a registry of object name to the query parameter used for performing incremental reading.
var incrementalQuery = map[string]string{ // nolint:gochecknoglobals
	"emails": "sentAtAfterOrEqualTo",
}

func (a *Adapter) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return Schemas.Select(objectNames)
}
