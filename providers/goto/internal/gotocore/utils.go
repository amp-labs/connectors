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

	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, path)
	if err != nil {
		return nil, fmt.Errorf("error building URL for object %s: %w", objectName, err)
	}

	url.WithQueryParam(queryParamSize, sampleSize)

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

	objectConfig, ok := objectRegistry[objectName]
	if !ok {
		return nil, fmt.Errorf("%w: no object config for %s", common.ErrMissingExpectedValues, objectName)
	}

	if objectConfig.service == serviceSCIM {
		records, ok := (*body)["resources"].([]any) // per SCIM spec, the array of records is always under the "Resources" key
		if !ok {
			return nil, fmt.Errorf("%w: SCIM response is missing Resources key", common.ErrMissingExpectedValues)
		}
		return records, nil
	}

	if objectConfig.service == serviceAdmin {
		records, ok := (*body)["results"].([]any) // per Admin API docs, the array of records is always under the "results" key
		if !ok {
			return nil, fmt.Errorf("%w: Admin API response is missing results key", common.ErrMissingExpectedValues)
		}
		return records, nil
	}

	embedded, ok := (*body)["_embedded"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: response is missing _embedded key", common.ErrMissingExpectedValues)
	}

	records, ok := embedded[objectName].([]any)
	if !ok {
		return nil, fmt.Errorf("%w: _embedded.%s is not an array", common.ErrMissingExpectedValues, objectName)
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
