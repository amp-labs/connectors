package goTo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

const (
	queryParamSize = "size"
	sampleSize     = "1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	// Sample a single record to infer the object's schema.
	url.WithQueryParam(queryParamSize, sampleSize)

	if objectName == "historicalMeetings" {
		startDate := time.Now().AddDate(0, 0, -60).Format(time.RFC3339) // set the start date to 2 months ago to ensure we get at least 1 record, as historical meetings are only available for the past 90 days
		endDate := time.Now().Format(time.RFC3339)
		url.WithQueryParam("startDate", startDate)
		url.WithQueryParam("endDate", endDate)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
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
