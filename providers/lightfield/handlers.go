package lightfield

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/lightfield/internal/metadata"
)

const (
	// Every request must carry the API version header.
	// https://docs.lightfield.app/using-the-api/api-keys
	versionHeaderName  = "Lightfield-Version"
	versionHeaderValue = "2026-03-01"

	// List endpoints accept a limit between 1 and 25.
	// https://docs.lightfield.app/using-the-api/list-endpoints#pagination
	maxPageSize = 25
)

var versionHeader = common.Header{ // nolint:gochecknoglobals
	Key:   versionHeaderName,
	Value: versionHeaderValue,
	Mode:  common.HeaderModeOverwrite,
}

// buildReadRequest lists records with limit/offset pagination.
// Lightfield has no updated-since filter, so Since/Until are applied connector side in parseReadResponse.
// TODO: The docs only allow filters on entries of `fields`, and updatedAt is a record property.
// If a live definitions response shows a filterable DATETIME update field, filter provider side instead.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	readURL, err := c.readURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
	if err != nil {
		return nil, err
	}

	versionHeader.ApplyToRequest(req)

	return req, nil
}

func (c *Connector) readURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam("limit", strconv.Itoa(min(readhelper.PageSizeWithDefault(params, maxPageSize), maxPageSize)))
	readURL.WithQueryParam("offset", "0")

	return readURL, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc("data"),
		makeUpdatedAtFilter(makeNextPageFunc(request.URL)),
		readhelper.MakeMarshaledDataFuncWithId(flattenRecord, readhelper.NewIdField("id")),
		params.Fields,
	)
}
