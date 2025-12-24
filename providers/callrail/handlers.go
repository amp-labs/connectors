package callrail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	restAPIVersion   = "v3/a/"
	limitQuery       = "per_page"
	metadataPageSize = "1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion+c.accountId, common.AddSuffixIfNotExists(objectName, ".json")) // nolint:lll
	if err != nil {
		return nil, err
	}

	// Limit response to 1 record data.
	// for all objects except calls
	if objectName != "calls" {
		url.WithQueryParam(limitQuery, metadataPageSize)
	}

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
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	fld := responseField.Get(objectName)

	records, okay := (*resp)[fld].([]any)
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
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (string, error) { // nolint: cyclop
	urlBuilder, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion+c.accountId, common.AddSuffixIfNotExists(params.ObjectName, ".json")) // nolint:lll
	if err != nil {
		return "", err
	}

	if params.NextPage != "" {
		urlBuilder.WithQueryParam("page", params.NextPage.String())
	}

	return urlBuilder.String(), nil
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
	var (
		paylod = params.RecordData
		method = http.MethodPost
		url    *urlbuilder.URL
		err    error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion+c.accountId, common.AddSuffixIfNotExists(params.ObjectName, ".json")) // nolint:lll
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion+c.accountId, params.ObjectName)
		if err != nil {
			return nil, err
		}

		url.AddPath(common.AddSuffixIfNotExists(params.RecordId, ".json"))

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(paylod)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func retrieveRecordId(body *ajson.Node) (string, error) {
	var idVal string

	// 1. we try integer
	recordID, err := jsonquery.New(body).IntegerWithDefault("id", 0)
	if !errors.Is(err, jsonquery.ErrNotNumeric) {
		return "", err
	}

	idVal = strconv.Itoa(int(recordID))

	// 2. we try string
	if recordID == 0 {
		recordId, err := jsonquery.New(body).StrWithDefault("id", "")
		if err != nil {
			return "nil", err
		}

		idVal = recordId
	}

	return idVal, nil
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

	recordID, err := retrieveRecordId(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		Data:     data,
		RecordId: recordID,
	}, nil
}
