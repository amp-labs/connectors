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
	limitQuery       = "limit"
	metadataPageSize = "1"
)

type dataResponse struct {
	Data []map[string]any `json:"data"`
	// Request Metadata Fields
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
