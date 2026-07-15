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

// buildIssueBody constructs the POST /issues/search payload. The search endpoint has no
// start_time/end_time parameters; a time window is instead expressed as a filter over
// updated_at, so that issues modified after the watermark are re-read rather than only
// newly created ones.
//
// filter is optional, and is omitted entirely when neither Since nor Until is set, which
// reads every issue.
func buildIssueBody(params common.ReadParams) (map[string]any, error) {
	body := map[string]any{
		"limit": searchLimit,
	}

	subfilters := make([]map[string]any, 0, 2) //nolint:gomnd

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
