// nolint:revive
package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSON(t *testing.T) { // nolint:funlen
	t.Parallel()

	type Story struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	tests := []struct {
		name        string
		input       []byte
		contentType string
		expected    *Story
		expectedErr error
	}{
		{
			name:        "Nil body produces empty struct",
			input:       nil,
			contentType: "application/json",
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:        "Empty body produces empty struct",
			input:       []byte(""),
			contentType: "application/json",
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid JSON produces marshal error",
			input:       []byte("2359"),
			contentType: "application/json",
			expected:    nil,
			expectedErr: ErrFailedToUnmarshalBody,
		},
		{
			name:        "Valid JSON values are mapped to struct fields",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Valid JSON with application/vnd.api+json content type",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/vnd.api+json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Valid JSON with application/schema+json content type",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/schema+json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid content type text/plain",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "text/plain",
			expected:    nil,
			expectedErr: ErrNotJSON,
		},
		{
			name:        "Invalid content type application/xml",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/xml",
			expected:    nil,
			expectedErr: ErrNotJSON,
		},
		{
			name:        "Empty content type, assume application/json and try to unmarshal",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
	}

	for _, ttc := range tests {
		// nolint:varnamelen
		t.Run(ttc.name, func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{
					"Content-Type": []string{ttc.contentType},
				},
				Body: io.NopCloser(bytes.NewReader(ttc.input)),
			}

			output, err := ParseJSONResponse(resp, ttc.input)
			if err != nil {
				testutils.CheckErrors(t, ttc.name, []error{ttc.expectedErr}, err)

				return
			}

			story, outErr := UnmarshalJSON[Story](output)
			testutils.CheckOutputWithError(t, ttc.name, ttc.expected, ttc.expectedErr, story, outErr)
		})
	}
}

// ============================================================================
// JSONHTTPCLIENT TESTS
// ============================================================================

func TestJSONHTTPClient_Get_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","data":{"id":"123"}}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	headers := []Header{{Key: "X-Custom", Value: "value"}}

	resp, err := client.Get(ctx, server.URL, headers...)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json", resp.Headers.Get("Content-Type"))

	body, ok := resp.Body()
	require.True(t, ok)

	status, err := body.GetKey("status")
	require.NoError(t, err)
	statusStr, err := status.GetString()
	require.NoError(t, err)
	assert.Equal(t, "ok", statusStr)

	data, err := body.GetKey("data")
	require.NoError(t, err)

	id, err := data.GetKey("id")
	require.NoError(t, err)
	idStr, err := id.GetString()
	require.NoError(t, err)
	assert.Equal(t, "123", idStr)
}

func TestJSONHTTPClient_Get_ErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	_, err := client.Get(ctx, server.URL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestJSONHTTPClient_Get_ErrorPostProcessor(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	customErr := errors.New("processed error")
	client := &JSONHTTPClient{
		HTTPClient: httpClient,
		ErrorPostProcessor: ErrorPostProcessor{
			Process: func(err error) error {
				return customErr
			},
		},
	}

	ctx := context.Background()
	_, err := client.Get(ctx, server.URL)

	require.ErrorIs(t, err, customErr)
}

func TestJSONHTTPClient_Post_Success(t *testing.T) {
	t.Parallel()

	expectedBody := map[string]string{"name": "test", "email": "test@example.com"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, expectedBody, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123","status":"created"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	headers := []Header{{Key: "X-Custom", Value: "value"}}

	resp, err := client.Post(ctx, server.URL, expectedBody, headers...)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "application/json", resp.Headers.Get("Content-Type"))

	body, ok := resp.Body()
	require.True(t, ok)

	id, err := body.GetKey("id")
	require.NoError(t, err)
	idStr, err := id.GetString()
	require.NoError(t, err)
	assert.Equal(t, "123", idStr)

	status, err := body.GetKey("status")
	require.NoError(t, err)
	statusStr, err := status.GetString()
	require.NoError(t, err)
	assert.Equal(t, "created", statusStr)
}

func TestJSONHTTPClient_Post_InvalidRequestBody(t *testing.T) {
	t.Parallel()

	httpClient := &HTTPClient{}
	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	// Create a body that cannot be marshaled to JSON
	invalidBody := make(chan int)

	_, err := client.Post(ctx, "https://example.com", invalidBody)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request body is not valid JSON")
}

func TestJSONHTTPClient_Put_Success(t *testing.T) {
	t.Parallel()

	expectedBody := map[string]string{"name": "updated", "email": "updated@example.com"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, expectedBody, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"123","status":"updated"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	resp, err := client.Put(ctx, server.URL, expectedBody)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, ok := resp.Body()
	require.True(t, ok)

	status, err := body.GetKey("status")
	require.NoError(t, err)
	statusStr, err := status.GetString()
	require.NoError(t, err)
	assert.Equal(t, "updated", statusStr)
}

func TestJSONHTTPClient_Patch_Success(t *testing.T) {
	t.Parallel()

	expectedBody := map[string]any{"op": "replace", "path": "/name", "value": "patched"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, expectedBody, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"patched"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	resp, err := client.Patch(ctx, server.URL, expectedBody)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, ok := resp.Body()
	require.True(t, ok)

	status, err := body.GetKey("status")
	require.NoError(t, err)
	statusStr, err := status.GetString()
	require.NoError(t, err)
	assert.Equal(t, "patched", statusStr)
}

func TestJSONHTTPClient_Delete_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	resp, err := client.Delete(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, "application/json", resp.Headers.Get("Content-Type"))

	// No content response should have empty body
	body, ok := resp.Body()
	assert.False(t, ok)
	assert.Nil(t, body)
	assert.Empty(t, resp.bodyBytes)
}

func TestJSONHTTPClient_Delete_WithResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"deleted"}`))
	}))
	defer server.Close()

	httpClient := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	client := &JSONHTTPClient{
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	resp, err := client.Delete(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, ok := resp.Body()
	require.True(t, ok)

	status, err := body.GetKey("status")
	require.NoError(t, err)
	statusStr, err := status.GetString()
	require.NoError(t, err)
	assert.Equal(t, "deleted", statusStr)
}

// ============================================================================
// JSONHTTPRESPONSE TESTS
// ============================================================================

func TestJSONHTTPResponse_Body(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response *JSONHTTPResponse
		wantBody bool
	}{
		{
			name: "response with body",
			response: &JSONHTTPResponse{
				bodyBytes: []byte(`{"test":"value"}`),
				body:      mustParseJSON(`{"test":"value"}`),
			},
			wantBody: true,
		},
		{
			name: "response with nil body",
			response: &JSONHTTPResponse{
				bodyBytes: []byte{},
				body:      nil,
			},
			wantBody: false,
		},
		{
			name: "response with empty body",
			response: &JSONHTTPResponse{
				bodyBytes: []byte(""),
				body:      nil,
			},
			wantBody: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body, ok := tt.response.Body()
			assert.Equal(t, tt.wantBody, ok)

			if tt.wantBody {
				assert.NotNil(t, body)
				// Verify we can get a value from the body
				value, err := body.GetKey("test")
				if tt.name == "response with body" {
					assert.NoError(t, err)
					valueStr, err := value.GetString()
					assert.NoError(t, err)
					assert.Equal(t, "value", valueStr)
				}
			} else {
				assert.Nil(t, body)
			}
		})
	}
}

func mustParseJSON(data string) *ajson.Node {
	node, err := ajson.Unmarshal([]byte(data))
	if err != nil {
		panic(err)
	}
	return node
}

// ============================================================================
// PARSEJSONRESPONSE TESTS
// ============================================================================

func TestParseJSONResponse_EmptyBody(t *testing.T) {
	t.Parallel()

	res := &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
	}

	response, err := ParseJSONResponse(res, []byte{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.bodyBytes)
	assert.Nil(t, response.body)

	body, ok := response.Body()
	assert.False(t, ok)
	assert.Nil(t, body)
}

func TestParseJSONResponse_ValidJSON(t *testing.T) {
	t.Parallel()

	bodyData := `{"name":"test","count":42,"active":true}`
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(bodyData)),
	}

	response, err := ParseJSONResponse(res, []byte(bodyData))

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, []byte(bodyData), response.bodyBytes)
	assert.NotNil(t, response.body)
	assert.Equal(t, "application/json", response.Headers.Get("Content-Type"))

	body, ok := response.Body()
	require.True(t, ok)

	name, err := body.GetKey("name")
	require.NoError(t, err)
	nameStr, err := name.GetString()
	require.NoError(t, err)
	assert.Equal(t, "test", nameStr)

	count, err := body.GetKey("count")
	require.NoError(t, err)
	countNum, err := count.GetNumeric()
	require.NoError(t, err)
	assert.Equal(t, 42.0, countNum)

	active, err := body.GetKey("active")
	require.NoError(t, err)
	activeBool, err := active.GetBool()
	require.NoError(t, err)
	assert.True(t, activeBool)
}

func TestParseJSONResponse_InvalidContentType(t *testing.T) {
	t.Parallel()

	bodyData := `{"name":"test"}`
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"text/plain"},
		},
		Body: io.NopCloser(strings.NewReader(bodyData)),
	}

	_, err := ParseJSONResponse(res, []byte(bodyData))

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotJSON)
	assert.Contains(t, err.Error(), "expected content type to be")
}

func TestParseJSONResponse_VariousJSONContentTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		expectError bool
	}{
		{
			name:        "standard application/json",
			contentType: "application/json",
			expectError: false,
		},
		{
			name:        "application/json with charset",
			contentType: "application/json; charset=utf-8",
			expectError: false,
		},
		{
			name:        "application/vnd.api+json",
			contentType: "application/vnd.api+json",
			expectError: false,
		},
		{
			name:        "application/schema+json",
			contentType: "application/schema+json",
			expectError: false,
		},
		{
			name:        "application/hal+json",
			contentType: "application/hal+json",
			expectError: false,
		},
		{
			name:        "text/plain (invalid)",
			contentType: "text/plain",
			expectError: true,
		},
		{
			name:        "application/xml (invalid)",
			contentType: "application/xml",
			expectError: true,
		},
		{
			name:        "empty content type",
			contentType: "",
			expectError: false, // Empty content type is allowed (no error)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bodyData := `{"test":"value"}`
			res := &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{tt.contentType},
				},
				Body: io.NopCloser(strings.NewReader(bodyData)),
			}

			response, err := ParseJSONResponse(res, []byte(bodyData))

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrNotJSON)
			} else {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, response.Code)

				if tt.contentType != "" {
					body, ok := response.Body()
					if ok {
						value, err := body.GetKey("test")
						assert.NoError(t, err)
						valueStr, err := value.GetString()
						assert.NoError(t, err)
						assert.Equal(t, "value", valueStr)
					}
				}
			}
		})
	}
}

func TestParseJSONResponse_InvalidJSON(t *testing.T) {
	t.Parallel()

	bodyData := `invalid json`
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(bodyData)),
	}

	_, err := ParseJSONResponse(res, []byte(bodyData))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshall response body into JSON")
}

func TestUnmarshalJSON_ComplexTypes(t *testing.T) {
	t.Parallel()

	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type User struct {
		ID      int      `json:"id"`
		Name    string   `json:"name"`
		Email   string   `json:"email"`
		Active  bool     `json:"active"`
		Tags    []string `json:"tags"`
		Address Address  `json:"address"`
	}

	jsonData := `{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com",
		"active": true,
		"tags": ["admin", "user"],
		"address": {
			"street": "123 Main St",
			"city": "Anytown",
			"country": "USA"
		}
	}`

	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(jsonData)),
	}

	response, err := ParseJSONResponse(res, []byte(jsonData))
	require.NoError(t, err)

	user, err := UnmarshalJSON[User](response)
	require.NoError(t, err)

	assert.Equal(t, 123, user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.True(t, user.Active)
	assert.Equal(t, []string{"admin", "user"}, user.Tags)
	assert.Equal(t, "123 Main St", user.Address.Street)
	assert.Equal(t, "Anytown", user.Address.City)
	assert.Equal(t, "USA", user.Address.Country)
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestAddAcceptJSONHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []Header
		expected []Header
	}{
		{
			name:     "nil headers",
			input:    nil,
			expected: []Header{{Key: "Accept", Value: "application/json"}},
		},
		{
			name:     "empty headers",
			input:    []Header{},
			expected: []Header{{Key: "Accept", Value: "application/json"}},
		},
		{
			name: "headers with Accept already present",
			input: []Header{
				{Key: "Accept", Value: "application/xml"},
				{Key: "Content-Type", Value: "application/json"},
			},
			expected: []Header{
				{Key: "Accept", Value: "application/xml"},
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Accept", Value: "application/json"}, // Appended
			},
		},
		{
			name: "headers without Accept",
			input: []Header{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Authorization", Value: "Bearer token"},
			},
			expected: []Header{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Authorization", Value: "Bearer token"},
				{Key: "Accept", Value: "application/json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := addAcceptJSONHeader(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakeJSONGetRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com/api/users"
	headers := []Header{
		{Key: "X-Custom", Value: "value"},
		{Key: "Authorization", Value: "Bearer token"},
	}

	req, err := MakeJSONGetRequest(ctx, url, headers)

	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
}

func TestAddSuffixIfNotExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		str      string
		suffix   string
		expected string
	}{
		{
			name:     "string already has suffix",
			str:      "test.json",
			suffix:   ".json",
			expected: "test.json",
		},
		{
			name:     "string doesn't have suffix",
			str:      "test",
			suffix:   ".json",
			expected: "test.json",
		},
		{
			name:     "empty string",
			str:      "",
			suffix:   ".json",
			expected: ".json",
		},
		{
			name:     "string with different suffix",
			str:      "test.xml",
			suffix:   ".json",
			expected: "test.xml.json",
		},
		{
			name:     "string ends with suffix but has extra",
			str:      "test.json.ext",
			suffix:   ".json",
			expected: "test.json.ext.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := AddSuffixIfNotExists(tt.str, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pattern       string
		contentType   string
		errOnMissing  bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid content type matches pattern",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "application/json",
			errOnMissing:  true,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "valid content type with charset",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "application/json; charset=utf-8",
			errOnMissing:  true,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "valid vendor-specific JSON",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "application/vnd.api+json",
			errOnMissing:  true,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "invalid content type",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "text/plain",
			errOnMissing:  true,
			expectError:   true,
			errorContains: "expected content type to be",
		},
		{
			name:          "missing content type, errOnMissing=true",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "",
			errOnMissing:  true,
			expectError:   true,
			errorContains: "missing content type",
		},
		{
			name:          "missing content type, errOnMissing=false",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "",
			errOnMissing:  false,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "malformed content type",
			pattern:       `^application/.*json([-0-9.])*$`,
			contentType:   "application/",
			errOnMissing:  true,
			expectError:   true,
			errorContains: "failed to parse content type",
		},
		{
			name:          "custom pattern matching",
			pattern:       `^text/(plain|html)$`,
			contentType:   "text/plain",
			errOnMissing:  true,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "custom pattern not matching",
			pattern:       `^text/(plain|html)$`,
			contentType:   "text/xml",
			errOnMissing:  true,
			expectError:   true,
			errorContains: "expected content type to be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{
				Header: http.Header{
					"Content-Type": []string{tt.contentType},
				},
			}

			err := EnsureContentType(tt.pattern, res, tt.errOnMissing)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnsureContentType_InvalidPattern(t *testing.T) {
	t.Parallel()

	res := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	// Invalid regex pattern
	err := EnsureContentType(`[invalid-regex`, res, true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compile regex")
}
