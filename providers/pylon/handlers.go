package pylon

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	url.WithQueryParam("limit", "1")

	if objectName == "issues" {
		// start_time and end_time is required for issues object
		url.WithQueryParam("start_time", time.Now().AddDate(0, 0, -30).Format(time.RFC3339))
		url.WithQueryParam("end_time", time.Now().Format(time.RFC3339))
	}

	if err != nil {
		return nil, err
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
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	records, ok := (*res)["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.ObjectName == "issues" {
		// Issues Object requires start_time and end_time query parameters
		// Time window should not exceed 30 days
		// If since is not provided, default to last 30 days

		var startTime, endTime time.Time

		if params.Since.IsZero() {
			startTime = time.Now().UTC().AddDate(0, 0, -30)
		} else {
			startTime = params.Since
		}

		if params.Until.IsZero() {
			endTime = time.Now().UTC()
		} else {
			endTime = params.Until
		}

		//Validate the time window does not exceed 30 days
		if endTime.Sub(startTime) > 30*24*time.Hour {
			return nil, fmt.Errorf("time window exceeds 30 days")
		}

		url.WithQueryParam("start_time", startTime.Format(time.RFC3339))
		url.WithQueryParam("end_time", endTime.Format(time.RFC3339))
	}

	if params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
