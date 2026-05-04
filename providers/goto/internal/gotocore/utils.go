package gotocore

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) buildObjectURL(objectName string) (*urlbuilder.URL, error) {
	spec, ok := objectRegistry[objectName]
	if !ok || spec.path == "" {
		spec.path = objectName
	}

	path := strings.ReplaceAll(spec.path, accountKeyPlaceholder, a.accountKey)

	return urlbuilder.New(a.ModuleInfo().BaseURL, path)
}

// extractRecords pulls the records array out of a GoTo response. Some
// endpoints (e.g. G2W) wrap results as {"_embedded": {"<recordsKey>": [...]}},
// while others (e.g. G2M historicalMeetings) return a top-level JSON array.
func extractRecords(response *common.JSONHTTPResponse, objectName string) ([]any, error) {
	if records, err := common.UnmarshalJSON[[]any](response); err == nil && records != nil {
		return *records, nil
	}

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
