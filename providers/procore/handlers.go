package procore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
)

const (
	defaultPageSize           = 1000
	headerProcoreCompanyID    = "Procore-Company-Id"
	queryParamPage            = "page"
	queryParamPerPage         = "per_page"
	queryParamUpdatedAtFilter = "filters[updated_at]"
	filterRangeSeparator      = "..."
)

var (
	ErrMissingCompanyID = errors.New("company metadata is required for this object")
	ErrInvalidObject    = errors.New("object name cannot be empty")
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(queryParamPerPage, "1")

	return c.newRequest(ctx, http.MethodGet, url)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName)),
	}

	records, err := extractRecords(response, objectName)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    analyzeValue(value),
			ProviderType: string(analyzeValue(value)),
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildObjectURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	page := "1"
	if params.NextPage != "" {
		page = params.NextPage.String()
	}

	url.WithQueryParam(queryParamPage, page)
	url.WithQueryParam(queryParamPerPage, strconv.Itoa(resolvePageSize(params.PageSize)))

	if objectRegistry[params.ObjectName].incremental {
		if filter := buildUpdatedAtFilter(params.Since, params.Until); filter != "" {
			url.WithQueryParam(queryParamUpdatedAtFilter, filter)
		}
	}

	return c.newRequest(ctx, http.MethodGet, url)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Procore paginates with a Link header, so we extract the next page token from there.
	linkHeader := response.Headers.Get("Link")

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(objectRegistry[params.ObjectName].recordsKey),
		nextPageFromLink(linkHeader),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}
