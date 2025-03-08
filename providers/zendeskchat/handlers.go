package zendeskchat

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	bans    = "bans"
	account = "account"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	firstRecord, err := parseIndividualResponse(objectName, response)
	if err != nil {
		return nil, err
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func parseIndividualResponse(objectName string, response *common.JSONHTTPResponse) (map[string]any, error) {
	field := responseFields(objectName)
	if field != "" {
		resp, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, err
		}

		records, ok := (*resp)[field].([]any)
		if !ok {
			return nil, fmt.Errorf("expected field '%s' to contain an array, got %T: %w", field, (*resp)[field], common.ErrMissingExpectedValues) //nolint:lll
		}

		if len(records) == 0 {
			return nil, fmt.Errorf("no records found in response: %w", common.ErrMissingExpectedValues)
		}

		firstRecord, ok := records[0].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected first record in field '%s' to be a map, got %T: %w", field, records[0], common.ErrMissingExpectedValues) //nolint:lll
		}

		return firstRecord, nil
	}

	switch objectName {
	case bans, account:
		return parseObjectResponse(response)
	default:
		return parseListResponse(response)
	}
}

func parseObjectResponse(response *common.JSONHTTPResponse) (map[string]any, error) {
	resp, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, err
	}

	return *resp, nil
}

func parseListResponse(response *common.JSONHTTPResponse) (map[string]any, error) {
	resp, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, err
	}

	if len(*resp) == 0 {
		return nil, fmt.Errorf("no records found in response: %w", common.ErrMissingExpectedValues)
	}

	return (*resp)[0], nil
}
