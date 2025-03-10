package monday

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	query := ""
	switch objectName {
	case "boards":
		query = fmt.Sprintf("query { boards { id name state permissions items_count columns { id title type } " +
			"groups { id title position } owner { id name } owners { id name } subscribers { id name } tags { id name } " +
			"team_owners { id name } team_subscribers { id name } top_group { id title } type updated_at " +
			"updates { id body created_at } url views { id name type } workspace { id name } workspace_id } }")
	case "users":
		query = "query { users { id email name enabled } }"
	default:
		return nil, fmt.Errorf("unsupported object name: %s", objectName)
	}

	// Create the request body as a map
	requestBody := map[string]string{
		"query": query,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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

	dataMap, ok := (*data)["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing data field")
	}

	rawRecords, exists := dataMap[objectName]
	if !exists {
		return nil, fmt.Errorf("missing expected values for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	records, ok := rawRecords.([]any)
	if len(records) == 0 || !ok {
		return nil, fmt.Errorf("unexpected type or empty records for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected record format for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
