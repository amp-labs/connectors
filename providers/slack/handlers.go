package slack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := objectName
	if !objectsWithoutListSuffix.Has(objectName) {
		urlPath = objectName + ".list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlPath)
	if err != nil {
		return nil, err
	}

	if objectsReadViaPost.Has(objectName) {
		return jsonPostRequest(ctx, url.String(), map[string]any{"limit": 1})
	}

	url.WithQueryParam("limit", "1")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateDotWords(objectName)),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrFailedToUnmarshalBody
	}

	recordsArr, err := getSlackResponseRecords(body, objectName)
	if err != nil {
		return nil, err
	}

	if len(recordsArr) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, err := jsonquery.Convertor.ObjectToMap(recordsArr[0])
	if err != nil {
		return nil, fmt.Errorf("couldn't convert the first record to an object: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "",
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	urlPath := params.ObjectName
	if !objectsWithoutListSuffix.Has(params.ObjectName) {
		urlPath = params.ObjectName + ".list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlPath)
	if err != nil {
		return nil, err
	}

	// Use 200 as the default page size, which is recommended by Slack.
	// Although Slack can support up to 1000, not all methods support the same limit,
	// and according to their docs this may change, so they recommend using 200.
	// Ref: https://docs.slack.dev/apis/web-api/pagination/
	pageSize := readhelper.PageSizeWithDefaultStr(params, "200")

	if objectsReadViaPost.Has(params.ObjectName) {
		body := map[string]any{"limit": pageSize}
		if params.NextPage != "" {
			body["cursor"] = params.NextPage.String()
		}

		return jsonPostRequest(ctx, url.String(), body)
	}

	url.WithQueryParam("limit", pageSize)

	if params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse( //nolint:unparam
	ctx context.Context, //nolint:revive
	params common.ReadParams,
	request *http.Request, //nolint:revive
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if objectsWithConnectorSideFilter.Has(params.ObjectName) {
		return common.ParseResultFiltered(
			params,
			response,
			nodeRecords(params.ObjectName),
			makeTimeFilter(params.ObjectName),
			readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
			params.Fields,
		)
	}

	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

// makeTimeFilter returns a RecordsFilterFunc that applies connector-side time filtering
// using the timestamp field for the given object. Slack does not support server-side
// date filtering, so we filter records in memory after fetching each page.
// Records are unordered, so pagination continues until all pages are exhausted.
func makeTimeFilter(objectName string) common.RecordsFilterFunc {
	timestampField := objectsWithConnectorSideFilter[objectName]

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		timestampField,
		readhelper.TimestampFormatUnixSec,
		nextRecordsURL(),
	)
}
