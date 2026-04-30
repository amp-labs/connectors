package goTo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildObjectURL(objectName string) (*urlbuilder.URL, error) {
	spec, ok := objectRegistry[objectName]
	if !ok || spec.path == "" {
		return nil, fmt.Errorf("%w: object %q is not registered", common.ErrOperationNotSupportedForObject, objectName)
	}

	path := strings.ReplaceAll(spec.path, organizerKeyPlaceholder, c.organizerKey)

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}

// extractRecords pulls the records array out of a GoTo Webinar response.
// Responses are wrapped as {"_embedded": {"<recordsKey>": [...]}, "page": {...}}.
func extractRecords(response *common.JSONHTTPResponse, objectName string) ([]any, error) {
	body, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil || body == nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	embedded, ok := (*body)["_embedded"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: response is missing _embedded key", common.ErrMissingExpectedValues)
	}

	recordsKey := objectRegistry[objectName].recordsKey
	if recordsKey == "" {
		recordsKey = objectName
	}

	records, ok := embedded[recordsKey].([]any)
	if !ok {
		return nil, fmt.Errorf("%w: _embedded.%s is not an array", common.ErrMissingExpectedValues, recordsKey)
	}

	return records, nil
}

func analyzeValue(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return common.ValueTypeOther
	}

	switch v.Kind() { //nolint:exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

//historicalMeetings
