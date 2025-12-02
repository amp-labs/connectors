package acuityscheduling

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
	apiVersion = "api/v1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("max", "1")

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
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	res, err := common.UnmarshalJSON[[]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := (*res)[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for fld, value := range firstRecord {
		objectMetadata.AddFieldMetadata(fld, common.FieldMetadata{
			DisplayName:  naming.CapitalizeFirstLetter(fld),
			ValueType:    analyzeValue(value),
			ProviderType: string(analyzeValue(value)),
		})
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("max", "100")

	if supportObjectIncrementalRead.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("minDate", params.Since.Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("maxDate", params.Until.Format(time.RFC3339))
		}
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
		common.ExtractRecordsFromPath(""),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
