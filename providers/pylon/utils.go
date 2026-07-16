package pylon

import (
	"time"

	"github.com/amp-labs/connectors/common"
)

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

// buildIssueBody constructs the POST /issues/search.
// We use a POST request because the GET /issues endpoint filters on created_at, which means
// issues updated after the watermark are never re-read. The POST /issues/search endpoint
// instead filters on updated_at, which is what we want.

// filter is optional, and is omitted entirely when neither Since nor Until is set, which
// reads every issue.
func buildIssueBody(params common.ReadParams) (map[string]any, error) {
	pageSize := searchLimit
	if params.PageSize > 0 && params.PageSize <= searchLimit {
		pageSize = params.PageSize
	}

	body := map[string]any{
		"limit": pageSize,
	}

	subfilters := make([]map[string]any, 0, 2) //nolint:gomnd,mnd

	if !params.Since.IsZero() {
		subfilters = append(subfilters, map[string]any{
			"field":    "updated_at",
			"operator": "time_is_after",
			"value":    params.Since.UTC().Format(time.RFC3339),
		})
	}

	if !params.Until.IsZero() {
		subfilters = append(subfilters, map[string]any{
			"field":    "updated_at",
			"operator": "time_is_before",
			"value":    params.Until.UTC().Format(time.RFC3339),
		})
	}

	if len(subfilters) > 0 {
		body["filter"] = map[string]any{
			"operator":   "and",
			"subfilters": subfilters,
		}
	}

	if params.NextPage != "" {
		body["cursor"] = params.NextPage.String()
	}

	return body, nil
}
