package amplitude

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiV2 = "2"
	apiV3 = "3"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

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

	responseField := objectResponseField.Get(objectName)

	data, ok := (*res)[responseField].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := data[0].(map[string]any)
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
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	var (
		url *urlbuilder.URL
		err error
	)

	url, err = c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := objectResponseField.Get(params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) { //nolint:lll
	method := http.MethodPost

	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()

	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: "",
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) handleAttributionObject(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	baseURL := "https://api2.amplitude.com/attribution"

	apiKey, err := c.extractAPIKey(ctx)
	if err != nil {
		return nil, err
	}

	eventJSON, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	form := make(url.Values)             // Creates an empty url.Values map
	form.Add("api_key", apiKey)          // Use the extracted API key
	form.Add("event", string(eventJSON)) // Add as JSON string

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

// Helper to extract API key (username) from the auth client
func (c *Connector) extractAPIKey(ctx context.Context) (string, error) {
	// Use a safe dummy endpoint (e.g., Amplitude's health check or base URL)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.amplitude.com/health", nil) // Adjust URL if needed
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	authHeader := resp.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no Authorization header found")
	}

	// Decode Basic auth: "Basic <base64(user:pass)>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", fmt.Errorf("invalid Basic auth header")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	creds := strings.Split(string(decoded), ":")
	if len(creds) != 2 {
		return "", fmt.Errorf("invalid credentials format")
	}

	return creds[0], nil // Return username (API key)
}
