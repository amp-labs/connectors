package lightfield

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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

	// DATETIME filters, `$<field>[<operator>]=<value>`. The `$` is part of the key:
	// `updatedAt[greaterThan]` is rejected with 400 unknown_field.
	// The operators match ReadParams: Since is exclusive and Until inclusive.
	// https://docs.lightfield.app/using-the-api/list-endpoints#filtering
	updatedSinceParam = "$updatedAt[greaterThan]"
	updatedUntilParam = "$updatedAt[lessThanOrEqual]"
)

var versionHeader = common.Header{ // nolint:gochecknoglobals
	Key:   versionHeaderName,
	Value: versionHeaderValue,
	Mode:  common.HeaderModeOverwrite,
}

// buildReadRequest lists records with limit/offset pagination. Since/Until are sent as
// `$updatedAt` filters where the endpoint accepts them; members are filtered in parseReadResponse.
// List results come from a search index that may lag recent changes.
// https://docs.lightfield.app/using-the-api/list-endpoints
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

	// updatedAt has millisecond precision, so the bounds keep milliseconds:
	// a whole-second Until would drop records updated within its last second.
	if objectsWithUpdatedAtFilter.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			readURL.WithQueryParam(updatedSinceParam, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
		}

		if !params.Until.IsZero() {
			readURL.WithQueryParam(updatedUntilParam, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Until))
		}
	}

	return readURL, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	nextPage := makeNextPageFunc(request.URL)

	filter := readhelper.MakeIdentityFilterFunc(nextPage)
	if !objectsWithUpdatedAtFilter.Has(params.ObjectName) {
		filter = makeUpdatedAtFilter(nextPage)
	}

	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc("data"),
		filter,
		readhelper.MakeMarshaledDataFuncWithId(flattenRecord, readhelper.NewIdField("id")),
		params.Fields,
	)
}
