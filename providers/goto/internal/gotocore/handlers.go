package gotocore

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

const (
	queryParamSize = "size"
	sampleSize     = "1"
)

func (a *Adapter) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := a.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	// Sample a single record to infer the object's schema.
	url.WithQueryParam(queryParamSize, sampleSize)

	if objectName == "historicalMeetings" {
		// Historical meetings are only available for the past 90 days, so we
		// query the last 120 days to ensure at least one record is returned.
		now := time.Now().UTC()
		startDate := now.AddDate(0, 0, -120).Format(time.RFC3339)
		endDate := now.Format(time.RFC3339)
		url.WithQueryParam("startDate", startDate)
		url.WithQueryParam("endDate", endDate)
	}

	if objectName == "webinars" {
		// webinars are only available for the past 90 days, so we
		// query the last 120 days to ensure at least one record is returned.
		now := time.Now().UTC()
		startDate := now.AddDate(0, 0, -120).Format(time.RFC3339)
		endDate := now.AddDate(0, 0, 120).Format(time.RFC3339)
		url.WithQueryParam("fromTime", startDate)
		url.WithQueryParam("toTime", endDate)
	}

	log.Printf("Building request for object %s with URL: %s", objectName, url.String())

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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
