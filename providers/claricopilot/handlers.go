package claricopilot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "limit"
	metadataPageSize = "1"
	pageSize         = "100"
	skipKey          = "skip"
	apiVersionV2     = "v2"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if supportedObjectV2.Has(objectName) {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersionV2, objectName)
	} else {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	url.WithQueryParam(limitQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	records, ok := (*data)[objectName].([]any)
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

	for field := range firstRecord {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if supportedObjectV2.Has(params.ObjectName) {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersionV2, params.ObjectName)
	} else {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	url.WithQueryParam(limitQuery, pageSize)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, fmt.Errorf("failed to build URL from next page: %w", err)
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
		common.ExtractRecordsFromPath(responseField(params.ObjectName)),
		nextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}
