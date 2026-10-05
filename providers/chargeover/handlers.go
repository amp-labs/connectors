package chargeover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiVersion    = "api/v3"
	responseField = "response"
	limitQuery    = "limit"
	offsetQuery   = "offset"
	pageSize      = 500
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	records, okay := (*resp)[responseField].([]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, okay := records[0].(map[string]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to a map: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	for fld, value := range firstRecord {
		objectMetadata.Fields[fld] = common.FieldMetadata{
			DisplayName:  fld,
			ValueType:    common.InferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

// escapeWhereValue formats a timestamp for use as a value inside ChargeOver's `where` query parameter.
// Colons and commas are reserved characters in a where clause and must be escaped with a backslash,
// see https://developer.chargeover.com/docs/api-information/escape-syntax/example-escape-syntax.
// Percent-encoding the colons instead makes ChargeOver stop parsing the timestamp at the first
// encoded colon, which silently truncates the filter to hour precision.
func escapeWhereValue(timestamp time.Time) string {
	return whereEscaper.Replace(timestamp.Format(time.RFC3339))
}

var whereEscaper = strings.NewReplacer(":", `\:`, ",", `\,`) //nolint:gochecknoglobals

type records struct {
	Response []map[string]any `json:"response"`
}

func (c *Connector) constructReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	path := params.ObjectName

	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	urlbuild, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	// some objects allows filtering in timestamp. If it does and the user provided
	// the required params, we use them.
	if !params.Since.IsZero() && !doNotFilter.Has(params.ObjectName) {
		if filteringFields.Has(params.ObjectName) {
			fld := filteringFields.Get(params.ObjectName)
			where := fld + ":GTE:" + escapeWhereValue(params.Since)

			if !params.Until.IsZero() {
				where += "," + fld + ":LTE:" + escapeWhereValue(params.Until)
			}

			urlbuild.WithQueryParam("where", where)
		}
	}

	// standard page size.
	urlbuild.WithQueryParam(limitQuery, strconv.Itoa(pageSize))

	return urlbuild, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	rec, err := common.UnmarshalJSON[records](response)
	if err != nil {
		return nil, err
	}

	numRecords := len(rec.Response)

	url, err := urlbuilder.New(request.URL.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(response,
		common.ExtractRecordsFromPath(responseField),
		nextRecordsURL(url, params.ObjectName, numRecords),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Errors:  nil,
		Data:    data,
	}, nil
}
