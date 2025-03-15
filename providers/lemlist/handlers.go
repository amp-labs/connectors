package lemlist

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("version", "v2")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	var (
		firstRecord map[string]any
		err         error
	)

	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	schema, fld := responseSchema(objectName)

	switch schema {
	case object:
		firstRecord, err = parseObject(response, fld)
		if err != nil {
			return nil, err
		}

	case list:
		firstRecord, err = parseList(response)
		if err != nil {
			return nil, err
		}
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func parseObject(response *common.JSONHTTPResponse, fld string) (map[string]any, error) {
	// We're unmarshaling the data to map[string]any,
	// all supported objects returns this data type.
	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord := *data

	if fld != "" {
		// If this is the case, we're expecting the data in a certain field
		// in this current map.
		records, okay := (*data)[fld].([]any)
		if !okay {
			return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
		}

		if len(records) == 0 {
			return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
		}

		// Iterate over the first record.
		firstRecord, okay = records[0].(map[string]any)
		if !okay {
			return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
		}
	}

	return firstRecord, nil
}

func parseList(response *common.JSONHTTPResponse) (map[string]any, error) {
	// We're unmarshaling the data to []map[string]any,
	// all supported objects returns this data type.
	data, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	return (*data)[0], nil
}
