package procore

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func resolveAPIPath(objectName, companyID string) string {
	spec, ok := objectRegistry[objectName]

	// If the object isn't registered, we fall back to a default path format.
	//  This allows for some flexibility in handling unregistered objects,
	//  but in general all supported objects should be registered with their correct paths.
	if !ok || spec.path == "" {
		return "rest/v1.0/" + objectName
	}

	return strings.ReplaceAll(spec.path, companyIDPlaceholder, companyID)
}

func (c *Connector) newRequest(ctx context.Context, method string, url *urlbuilder.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerProcoreCompanyID, c.companyId)

	return req, nil
}

func (c *Connector) buildObjectURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, resolveAPIPath(objectName, c.companyId))
}

func resolvePageSize(requested int) int {
	if requested > 0 {
		return requested
	}

	return defaultPageSize
}

// buildUpdatedAtFilter encodes since/until as Procore's filters[updated_at] range.
// Procore accepts a half-open form like `<since>...<until>`.
func buildUpdatedAtFilter(since, until time.Time) string {
	if since.IsZero() && until.IsZero() {
		return ""
	}

	var s, u string
	if !since.IsZero() {
		s = since.UTC().Format(time.RFC3339)
	}

	if !until.IsZero() {
		u = until.UTC().Format(time.RFC3339)
	}

	return s + filterRangeSeparator + u
}

func analyzeValue(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	v := reflect.ValueOf(value)

	if !v.IsValid() {
		return common.ValueTypeOther
	}

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
