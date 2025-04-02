package monday

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	// ErrUnsupportedObject is returned when an unsupported object type is requested.
	ErrUnsupportedObject = errors.New("unsupported object")
	// ErrInvalidResponseFormat is returned when the API response format is unexpected.
	ErrInvalidResponseFormat = errors.New("invalid response format")
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	var query string

	switch objectName {
	case "boards":
		query = `query {
			boards {
				id name state permissions items_count
				columns { id title type }
				groups { id title position }
				owner { id name }
				owners { id name }
				subscribers { id name }
				tags { id name }
				team_owners { id name }
				team_subscribers { id name }
				top_group { id title }
				type updated_at
				updates { id body created_at }
				url
				views { id name type }
				workspace { id name }
				workspace_id
			}
		}`
	case "users":
		query = `query { users { id email name enabled } }`
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedObject, objectName)
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

	dataMap, isValidData := (*data)["data"].(map[string]any)
	if !isValidData {
		return nil, fmt.Errorf("%w: missing data field", ErrInvalidResponseFormat)
	}

	rawRecords, exists := dataMap[objectName]
	if !exists {
		return nil, fmt.Errorf("missing expected values for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	records, isValidRecords := rawRecords.([]any)
	if len(records) == 0 || !isValidRecords {
		return nil, fmt.Errorf("unexpected type or empty records for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	firstRecord, isValidRecord := records[0].(map[string]any)
	if !isValidRecord {
		return nil, fmt.Errorf("unexpected record format for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
