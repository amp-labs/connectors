package hunter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	limitQuery       = "limit"
	metadataPageSize = "1"
	offsetKey        = "offset"
	totalField       = "total"
	paramsField      = "params"
	metaField        = "meta"
	pageSize         = 100
)

type readResponse struct {
	Data map[string]any `json:"data"`
	Meta map[string]any `json:"meta"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.RootClient.URL(restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	// Limit response to 1 record data.
	url.WithQueryParam(limitQuery, metadataPageSize)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[readResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	records, okay := resp.Data[objectName].([]any)
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

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = c.RootClient.URL(restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(limitQuery, strconv.Itoa(pageSize))

	if params.NextPage != "" {
		url.WithQueryParam(offsetKey, params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := c.RootClient.URL(restAPIVersion, params.ObjectName)
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

	resp, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	recordId, err := jsonquery.New(resp).IntegerOptional("id")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: strconv.Itoa(int(*recordId)),
		Data:     data,
	}, nil
}
