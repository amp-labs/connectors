package gorgias

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	account          = "account"
	dataField        = "data"
	idField          = "id"
	limitQuery       = "limit"
	cursorQuery      = "cursor"
	metadataPageSize = "1"
	readPageSize     = "100"
)

type dataResponse struct {
	Data []map[string]any `json:"data"`
	Meta map[string]any   `json:"meta"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, objectName)
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
	objectMetadata := common.NewObjectMetadata(
		naming.CapitalizeFirstLetterEveryWord(objectName),
		common.FieldsMetadata{},
	)

	// All supported objects return a response following the `dataResponse` schema,
	// with the exception of the `account` object.
	switch objectName {
	case account:
		record, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		for fld := range *record {
			objectMetadata.AddField(fld, fld)
		}
	default:
		records, err := common.UnmarshalJSON[dataResponse](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		if len(records.Data) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		for fld := range records.Data[0] {
			objectMetadata.AddField(fld, fld)
		}
	}

	return objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// request the maximum allowed retrieval records per request.
	url.WithQueryParam(limitQuery, readPageSize)

	if params.NextPage != "" {
		url.WithQueryParam(cursorQuery, params.NextPage.String())
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
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
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
	ctx = logging.With(ctx, "gorgias")

	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	var recordIDStr string

	recordID, err := jsonquery.New(body).IntegerOptional(idField)
	if err != nil {
		logging.Logger(ctx).Error("failed to retrieve the ID from the response", "error", err.Error())
	}

	if recordID != nil {
		recordIDStr = strconv.Itoa(int(*recordID))
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		logging.Logger(ctx).Error("failed to convert the response to a map", "error", err.Error())
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordIDStr,
		Errors:   nil,
		Data:     data,
	}, nil
}
