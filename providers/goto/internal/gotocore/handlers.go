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

const (
	queryParamSize     = "size"
	queryParamPageSize = "pageSize"
	sampleSize         = "1"

	// metadataSampleWindowDays is the size in days of the time-range filter
	// applied when sampling records for schema. Wide enough to
	// catch at least one record on endpoints that mandate a
	// time-range filter.
	metadataSampleWindowDays = 400
)

func (a *Adapter) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := a.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	applyMetadataTimeFilter(url, objectName)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// applyMetadataTimeFilter adds the mandatory time-range query params for
// endpoints that require them. The window is wide enough (past 120 days,
// plus 120 days into the future for endpoints that accept upcoming records)
// to maximize the chance of sampling at least one record for schema
// inference.
func applyMetadataTimeFilter(url *urlbuilder.URL, objectName string) {
	setWindow := func(startParam, endParam string, pastDays, futureDays int) {
		now := time.Now().UTC()
		url.WithQueryParam(startParam, now.AddDate(0, 0, -pastDays).Format(time.RFC3339))
		url.WithQueryParam(endParam, now.AddDate(0, 0, futureDays).Format(time.RFC3339))
	}

	switch objectName {
	case "historicalMeetings":
		setWindow("startDate", "endDate", metadataSampleWindowDays, 0)
	case "webinars":
		setWindow("fromTime", "toTime", metadataSampleWindowDays, metadataSampleWindowDays)
	case "sessions":
		setWindow("fromTime", "toTime", metadataSampleWindowDays, 0)
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
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateCamelCaseWords(objectName)),
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
