package outplay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	apiVersion = "v1"
	timeLayout = "2006-01-02"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	apiPath := objectAPIPath.Get(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, apiPath)
	if err != nil {
		return nil, err
	}

	if objectName == ObjectNameProspectMails || objectName == ObjectNameCallAnalysis {
		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	body, err := buildMetadataBody(objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), body)
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

	records, err := extractMetadataRecords(*res, objectName)
	if err != nil {
		return nil, err
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
	apiPath := objectAPIPath.Get(params.ObjectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, apiPath)
	if err != nil {
		return nil, err
	}

	// prospectmails and callanalysis use GET method for read
	if params.ObjectName == ObjectNameProspectMails || params.ObjectName == ObjectNameCallAnalysis {
		buildReadQueryParams(url, params)

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	body, err := buildReadBody(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), body)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// callanalysis has a different response structure
	// data is nested inside another data field
	// https://documenter.getpostman.com/view/16947449/TzsikPV1#c44371cf-0819-4eb9-805a-7fdde9b4f9dc
	if params.ObjectName == ObjectNameCallAnalysis {
		return common.ParseResult(
			response,
			common.ExtractRecordsFromPath("data", "data"),
			nextRecordsURLForCallAnalysis(),
			common.GetMarshaledData,
			params.Fields,
		)
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
