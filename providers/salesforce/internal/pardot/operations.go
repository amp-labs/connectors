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

	rsp, err := a.Client.Get(ctx, url.String(), a.businessUnitHeader())
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
	objectNameLower := strings.ToLower(params.ObjectName)

	url, err := a.getURL(objectNameLower)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	if len(params.RecordId) == 0 {
		// Create.
		write = a.Client.Post
	} else {
		// Update.
		write = a.Client.Patch

		url.AddPath(params.RecordId)
	}

	res, err := write(ctx, url.String(), params.RecordData, a.businessUnitHeader())
	if err != nil {
		return nil, err
	}

	return constructWriteResult(res)
}

func constructWriteResult(res *common.JSONHTTPResponse) (*common.WriteResult, error) {
	body, ok := res.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (a *Adapter) Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error) {
	objectNameLower := strings.ToLower(params.ObjectName)

	url, err := a.getURL(objectNameLower)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	// 204 "No Content" is expected
	_, err = a.Client.Delete(ctx, url.String(), a.businessUnitHeader())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return Schemas.Select(objectNames)
}
