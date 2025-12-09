package pylon

import (
	"errors"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

func addIssuesTimeWindowQuery(url *urlbuilder.URL, params common.ReadParams) error {
	var startTime, endTime time.Time

	if params.Since.IsZero() {
		// Default to last 29 days instead of 30 to account for potential
		// timezone transitions and microsecond precision differences that
		// could cause the API to reject requests exceeding 30 days exactly.
		startTime = time.Now().UTC().AddDate(0, 0, -29)
	} else {
		startTime = params.Since.UTC()
	}

	if params.Until.IsZero() {
		endTime = time.Now().UTC()
	} else {
		endTime = params.Until.UTC()
	}

	if endTime.Sub(startTime) > 30*24*time.Hour {
		return errors.New("time window exceeds 30 days") // nolint:err113
	}

	url.WithQueryParam("start_time", startTime.Format(time.RFC3339))
	url.WithQueryParam("end_time", endTime.Format(time.RFC3339))

	return nil
}
