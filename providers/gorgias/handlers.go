package gorgias

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	account          = "account"
	dataField        = "data"
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
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	// All supported objects return a response following the `dataResponse` schema,
	// with the exception of the `account` object.
	switch objectName {
	case account:
		record, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		for fld := range *record {
			objectMetadata.FieldsMap[fld] = fld
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
			objectMetadata.FieldsMap[fld] = fld
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// request the maximum allowed retrieval records per request.
	url.WithQueryParam(limitQuery, "10")

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
