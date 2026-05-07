package gotocore

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) buildObjectURL(objectName string) (*urlbuilder.URL, error) {
	url, err := a.buildObjectBaseURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(queryParamSize, sampleSize)

	return url, nil
}

// buildObjectBaseURL resolves the object path against the module BaseURL,
// substituting the account key. It does not attach any query params.
func (a *Adapter) buildObjectBaseURL(objectName string) (*urlbuilder.URL, error) {
	spec, ok := objectRegistry[objectName]
	if !ok || spec.path == "" {
		spec.path = objectName
	}

	path := strings.ReplaceAll(spec.path, accountKeyPlaceholder, a.accountKey)

	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, path)
	if err != nil {
		return nil, fmt.Errorf("error building URL for object %s: %w", objectName, err)
	}

	return url, nil
}

// extractRecords pulls the records array out of a GoTo response. Response
// shapes vary by service: SCIM wraps under "resources", Admin under
// "results", G2W under "_embedded.<objectName>", and some endpoints (e.g.
// G2M historicalMeetings) return a bare top-level JSON array.
func extractRecords(response *common.JSONHTTPResponse, objectName string) ([]any, error) {
	if records, err := common.UnmarshalJSON[[]any](response); err == nil && records != nil {
		return *records, nil
	}

	body, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil || body == nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	cfg, ok := objectRegistry[objectName]
	if !ok {
		return nil, fmt.Errorf("%w: no object config for %s", common.ErrMissingExpectedValues, objectName)
	}

	return extractRecordsByService(*body, cfg.service, objectName)
}

func extractRecordsByService(body map[string]any, service objectService, objectName string) ([]any, error) {
	switch service { //nolint:exhaustive // _embedded shape is the default; remaining services fall through.
	case serviceSCIM:
		return readArrayKey(body, "resources", objectName)
	case serviceAdmin:
		return readArrayKey(body, "results", objectName)
	case serviceAssist:
		return readArrayKey(body, objectName, objectName)
	default:
		// Webinar or any future GoTo services that
		// share the standard HAL-style envelope return records under
		// _embedded.<objectName>.
		embedded, ok := body["_embedded"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: unrecognized response shape for object %s",
				common.ErrMissingExpectedValues, objectName)
		}

		return readArrayKey(embedded, objectName, objectName)
	}
}

func readArrayKey(m map[string]any, key, objectName string) ([]any, error) {
	records, ok := m[key].([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s response is missing %q key", common.ErrMissingExpectedValues, objectName, key)
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
