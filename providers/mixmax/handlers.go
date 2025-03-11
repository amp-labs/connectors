package mixmax

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "per_page"
	metadataPageSize = "1"
	pageSize         = "50"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
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

	// We're unmarshaling the data to map[string]any,
	// all supported objects returns this data type.
	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	field := responseField(objectName)
	firstRecord := *data

	// checks if the response data field is available, else assumes we are in the data object.
	if field != "" {
		// If this is the case, we're expecting the data in a certain field
		// in this current map.
		records, ok := (*data)[field].([]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
		}

		// Iterate over the first record.
		firstRecord, ok = records[0].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
		}
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
