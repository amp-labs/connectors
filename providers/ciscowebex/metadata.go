package ciscowebex

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Most objects use "items", but some use different keys (e.g., "groups")
var objectResponseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"groups": "groups",
}, func(key string) string {
	return "items" // default response key for most Webex objects
})

// Groups uses "count", others use "max"
var objectLimitQueryParam = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"groups": "count",
}, func(key string) string {
	return "max" // default query param for most Webex objects
})

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v1", objectName)
	if err != nil {
		return nil, err
	}

	limitParam := objectLimitQueryParam.Get(objectName)
	url.WithQueryParam(limitParam, "1")

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
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	responseKey := objectResponseField.Get(objectName)

	records, ok := (*res)[responseKey].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the %s response field to an array: %w", responseKey, common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
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

//nolint:unused // Used in parseSingleObjectMetadataResponse
func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean

	default:
		return common.ValueTypeOther
	}
}
