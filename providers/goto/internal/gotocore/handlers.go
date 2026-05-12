package gotocore

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := a.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	applyTimeFilter(url, objectName, time.Time{}, time.Time{})

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// applyTimeFilter adds the mandatory time-range query params for endpoints
// that require them. Caller-supplied since/until win when non-zero;
// otherwise a default window (past 120 days, plus 120 days into the future
// for endpoints that accept upcoming records) is used so that metadata
// sampling and unbounded reads still hit at least one record.
func applyTimeFilter(url *urlbuilder.URL, objectName string, since, until time.Time) {
	now := time.Now().UTC()
	pick := func(t, fallback time.Time) time.Time {
		if t.IsZero() {
			return fallback
		}

		return t
	}

	setWindow := func(startParam, endParam string, defStart, defEnd time.Time) {
		url.WithQueryParam(startParam, pick(since, defStart).Format(time.RFC3339))
		url.WithQueryParam(endParam, pick(until, defEnd).Format(time.RFC3339))
	}

	defaultPast := now.AddDate(0, 0, -metadataSampleWindowDays)
	defaultFuture := now.AddDate(0, 0, metadataSampleWindowDays)

	switch objectName {
	case "historicalMeetings":
		setWindow("startDate", "endDate", defaultPast, now)
	case "webinars":
		setWindow("fromTime", "toTime", defaultPast, defaultFuture)
	case "sessions":
		setWindow("fromTime", "toTime", defaultPast, now)
	}
}

func (a *Adapter) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName)),
	}

	records, err := extractRecords(response, objectName)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		valueType := analyzeValue(value)
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    valueType,
			ProviderType: string(valueType),
		}
	}

	return &objectMetadata, nil
}
