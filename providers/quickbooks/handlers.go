package quickbooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	c.realmID = "9341455309256114" // QuickBooks Company ID, should be set dynamically

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, c.realmID, "query")
	if err != nil {
		return nil, err
	}

	Query := "SELECT * FROM " + naming.CapitalizeFirstLetter(objectName) + " STARTPOSITION 0 MAXRESULTS 1"

	url.WithQueryParam("query", Query)

	httpClient, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient.Header.Set("Accept", "application/json")

	return httpClient, nil
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

	QueryResponse, ok := (*res)["QueryResponse"].(map[string]any) // nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field QueryResponse to a map: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	records, ok := QueryResponse[naming.CapitalizeFirstLetter(objectName)].([]any) // nolint:varnamelen

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
