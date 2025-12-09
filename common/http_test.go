// nolint:revive
package common

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyTransport struct{}

type testKey string

func (d *dummyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	rsp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("hello")),
	}

	for k, v1 := range request.Header {
		for _, v2 := range v1 {
			rsp.Header.Add(k, v2)
		}
	}

	return rsp, nil
}

func TestModifierInClient(t *testing.T) {
	t.Parallel()

	modifier, ok := getRequestModifier(t.Context())
	require.False(t, ok)
	require.Nil(t, modifier)

	client := &http.Client{
		Transport: &dummyTransport{},
	}

	ac, err := NewHeaderAuthHTTPClient(t.Context(), WithHeaderClient(client))
	require.NoError(t, err)

	ctx := WithRequestModifier(t.Context(), func(req *http.Request) {
		req.Header.Set("Header", "value")
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	rsp, err := ac.Do(req)
	require.NoError(t, err)

	defer func() {
		_ = rsp.Body.Close()
	}()

	vals := rsp.Header.Values("Header")
	require.Len(t, vals, 1)

	require.Equal(t, "value", vals[0])
}

func TestModifierStandalone(t *testing.T) {
	t.Parallel()

	modifier, ok := getRequestModifier(t.Context())
	require.False(t, ok)
	require.Nil(t, modifier)

	ctx := WithRequestModifier(t.Context(), func(req *http.Request) {
		req.Header.Set("Header", "value")
	})

	modifier, ok = getRequestModifier(ctx)
	require.True(t, ok)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	modifier(req)

	require.Equal(t, "value", req.Header.Get("Header"))
}

func TestApplySetIfMissingEmptyHeadersRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	headers := Headers{
		{
			Key:   "Header",
			Value: "value2",
			Mode:  HeaderModeSetIfMissing,
		},
	}

	headers.ApplyToRequest(req)

	vals := req.Header.Values("Header")
	require.Len(t, vals, 1)

	require.Equal(t, "value2", vals[0])
}

func TestApplySetIfMissingNonEmptyHeadersRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("Header", "value1")

	headers := Headers{
		{
			Key:   "Header",
			Value: "value2",
			Mode:  HeaderModeSetIfMissing,
		},
	}

	headers.ApplyToRequest(req)

	vals := req.Header.Values("Header")
	require.Len(t, vals, 1)

	require.Equal(t, "value1", vals[0])
}

func TestApplySetHeadersEmptyRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	headers := Headers{
		{
			Key:   "Header",
			Value: "value2",
			Mode:  HeaderModeOverwrite,
		},
	}

	headers.ApplyToRequest(req)

	vals := req.Header.Values("Header")
	require.Len(t, vals, 1)

	require.Equal(t, "value2", vals[0])
}

func TestApplySetHeadersNonEmptyRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("Header", "value1")

	headers := Headers{
		{
			Key:   "Header",
			Value: "value2",
			Mode:  HeaderModeOverwrite,
		},
	}

	headers.ApplyToRequest(req)

	vals := req.Header.Values("Header")
	require.Len(t, vals, 1)

	require.Equal(t, "value2", vals[0])
}

func TestApplyAppendHeadersToRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("Header", "value1")

	headers := Headers{
		{
			Key:   "Header",
			Value: "value2",
			Mode:  HeaderModeAppend,
		},
		{
			Key:   "Header",
			Value: "value3",
		},
	}

	headers.ApplyToRequest(req)

	vals := req.Header.Values("Header")
	require.Len(t, vals, 3)

	require.Equal(t, "value1", vals[0])
	require.Equal(t, "value2", vals[1])
	require.Equal(t, "value3", vals[2])
}

type mockAuthClient struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (m *mockAuthClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("no mock function provided")
}

func (m *mockAuthClient) CloseIdleConnections() {}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// ============================================================================
// HEADER TESTS
// ============================================================================

func TestHeader_ApplyToRequest_Append(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("X-Custom", "value1")

	header := Header{
		Key:   "X-Custom",
		Value: "value2",
		Mode:  HeaderModeAppend,
	}

	header.ApplyToRequest(req)

	vals := req.Header.Values("X-Custom")
	assert.Len(t, vals, 2)
	assert.Equal(t, "value1", vals[0])
	assert.Equal(t, "value2", vals[1])
}

func TestHeader_ApplyToRequest_Overwrite(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("X-Custom", "value1")
	req.Header.Add("X-Custom", "value2")

	header := Header{
		Key:   "X-Custom",
		Value: "value3",
		Mode:  HeaderModeOverwrite,
	}

	header.ApplyToRequest(req)

	vals := req.Header.Values("X-Custom")
	assert.Len(t, vals, 1)
	assert.Equal(t, "value3", vals[0])
}

func TestHeader_ApplyToRequest_SetIfMissing_WhenMissing(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	header := Header{
		Key:   "X-Custom",
		Value: "value1",
		Mode:  HeaderModeSetIfMissing,
	}

	header.ApplyToRequest(req)

	vals := req.Header.Values("X-Custom")
	assert.Len(t, vals, 1)
	assert.Equal(t, "value1", vals[0])
}

func TestHeader_ApplyToRequest_SetIfMissing_WhenPresent(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("X-Custom", "existing")

	header := Header{
		Key:   "X-Custom",
		Value: "new",
		Mode:  HeaderModeSetIfMissing,
	}

	header.ApplyToRequest(req)

	vals := req.Header.Values("X-Custom")
	assert.Len(t, vals, 1)
	assert.Equal(t, "existing", vals[0])
}

func TestHeader_ApplyToRequest_DefaultMode(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	req.Header.Add("X-Custom", "value1")

	header := Header{
		Key:   "X-Custom",
		Value: "value2",
		Mode:  headerModeUnset,
	}

	header.ApplyToRequest(req)

	vals := req.Header.Values("X-Custom")
	assert.Len(t, vals, 2)
	assert.Equal(t, "value1", vals[0])
	assert.Equal(t, "value2", vals[1])
}

func TestHeader_String(t *testing.T) {
	t.Parallel()

	header := Header{
		Key:   "Content-Type",
		Value: "application/json",
	}

	assert.Equal(t, "Content-Type: application/json", header.String())
}

func TestHeader_Equals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		h1       Header
		h2       Header
		expected bool
	}{
		{
			name:     "identical headers",
			h1:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			h2:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			expected: true,
		},
		{
			name:     "case insensitive key",
			h1:       Header{Key: "content-type", Value: "application/json", Mode: HeaderModeAppend},
			h2:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			expected: true,
		},
		{
			name:     "different values",
			h1:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			h2:       Header{Key: "Content-Type", Value: "text/plain", Mode: HeaderModeAppend},
			expected: false,
		},
		{
			name:     "different modes",
			h1:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			h2:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeOverwrite},
			expected: false,
		},
		{
			name:     "different keys",
			h1:       Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
			h2:       Header{Key: "Accept", Value: "application/json", Mode: HeaderModeAppend},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.h1.equals(tt.h2))
		})
	}
}

func TestHeaders_Has(t *testing.T) {
	t.Parallel()

	headers := Headers{
		{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
		{Key: "Authorization", Value: "Bearer token", Mode: HeaderModeAppend},
	}

	assert.True(t, headers.Has(Header{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend}))
	assert.True(t, headers.Has(Header{Key: "content-type", Value: "application/json", Mode: HeaderModeAppend}))
	assert.False(t, headers.Has(Header{Key: "Accept", Value: "application/json", Mode: HeaderModeAppend}))
	assert.False(t, headers.Has(Header{Key: "Content-Type", Value: "text/plain", Mode: HeaderModeAppend}))
}

func TestHeaders_ApplyToRequest_MultipleHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	headers := Headers{
		{Key: "Content-Type", Value: "application/json", Mode: HeaderModeAppend},
		{Key: "Authorization", Value: "Bearer token", Mode: HeaderModeAppend},
		{Key: "X-Custom", Value: "value", Mode: HeaderModeAppend},
	}

	headers.ApplyToRequest(req)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
}

func TestHeaders_LogValue(t *testing.T) {
	t.Parallel()

	headers := Headers{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Authorization", Value: "Bearer token"},
	}

	logValue := headers.LogValue()
	assert.NotNil(t, logValue)
}

func TestRedactSensitiveRequestHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		headers  []Header
		expected Headers
	}{
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
		{
			name: "redact Authorization",
			headers: []Header{
				{Key: "Authorization", Value: "Bearer secret-token"},
				{Key: "Content-Type", Value: "application/json"},
			},
			expected: Headers{
				{Key: "Authorization", Value: "<redacted>"},
				{Key: "Content-Type", Value: "application/json"},
			},
		},
		{
			name: "redact Proxy-Authorization",
			headers: []Header{
				{Key: "Proxy-Authorization", Value: "Basic secret"},
			},
			expected: Headers{
				{Key: "Proxy-Authorization", Value: "<redacted>"},
			},
		},
		{
			name: "redact x-amz-security-token",
			headers: []Header{
				{Key: "x-amz-security-token", Value: "aws-secret"},
			},
			expected: Headers{
				{Key: "x-amz-security-token", Value: "<redacted>"},
			},
		},
		{
			name: "redact X-Api-Key",
			headers: []Header{
				{Key: "X-Api-Key", Value: "api-key-123"},
			},
			expected: Headers{
				{Key: "X-Api-Key", Value: "<redacted>"},
			},
		},
		{
			name: "redact X-Admin-Key",
			headers: []Header{
				{Key: "X-Admin-Key", Value: "admin-key-456"},
			},
			expected: Headers{
				{Key: "X-Admin-Key", Value: "<redacted>"},
			},
		},
		{
			name: "case insensitive redaction",
			headers: []Header{
				{Key: "authorization", Value: "secret"},
				{Key: "AUTHORIZATION", Value: "secret"},
			},
			expected: Headers{
				{Key: "authorization", Value: "<redacted>"},
				{Key: "AUTHORIZATION", Value: "<redacted>"},
			},
		},
		{
			name: "mixed headers",
			headers: []Header{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Authorization", Value: "secret"},
				{Key: "X-Request-ID", Value: "123"},
				{Key: "X-Api-Key", Value: "key"},
			},
			expected: Headers{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Authorization", Value: "<redacted>"},
				{Key: "X-Request-ID", Value: "123"},
				{Key: "X-Api-Key", Value: "<redacted>"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := redactSensitiveRequestHeaders(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactSensitiveResponseHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		headers  []Header
		expected Headers
	}{
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
		{
			name: "redact Set-Cookie",
			headers: []Header{
				{Key: "Set-Cookie", Value: "session=secret"},
				{Key: "Content-Type", Value: "application/json"},
			},
			expected: Headers{
				{Key: "Set-Cookie", Value: "<redacted>"},
				{Key: "Content-Type", Value: "application/json"},
			},
		},
		{
			name: "case insensitive",
			headers: []Header{
				{Key: "set-cookie", Value: "secret"},
				{Key: "SET-COOKIE", Value: "secret"},
			},
			expected: Headers{
				{Key: "set-cookie", Value: "<redacted>"},
				{Key: "SET-COOKIE", Value: "<redacted>"},
			},
		},
		{
			name: "non-sensitive headers",
			headers: []Header{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "X-Request-ID", Value: "123"},
			},
			expected: Headers{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "X-Request-ID", Value: "123"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := redactSensitiveResponseHeaders(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// HTTPCLIENT TESTS
// ============================================================================

func TestHTTPClient_Get_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	headers := []Header{{Key: "Accept", Value: "application/json"}}

	resp, body, err := client.Get(ctx, server.URL, headers...)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"status":"ok"}`, string(body))
}

func TestHTTPClient_Get_WithBaseURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/users", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &HTTPClient{
		Base: server.URL,
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	resp, _, err := client.Get(ctx, "/api/users")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClient_Get_ErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	_, _, err := client.Get(ctx, server.URL)

	require.Error(t, err)
}

func TestHTTPClient_Get_CustomErrorHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	customErr := errors.New("custom error")
	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ErrorHandler: func(rsp *http.Response, body []byte) error {
			return customErr
		},
	}

	ctx := context.Background()
	_, _, err := client.Get(ctx, server.URL)

	require.ErrorIs(t, err, customErr)
}

func TestHTTPClient_Post_Success(t *testing.T) {
	t.Parallel()

	expectedBody := `{"name":"test"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, expectedBody, string(body))

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	resp, body, err := client.Post(ctx, server.URL, []byte(expectedBody))

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.JSONEq(t, `{"id":"123"}`, string(body))
}

func TestHTTPClient_Post_URLEncoded(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		assert.Contains(t, bodyStr, "key1=value1")
		assert.Contains(t, bodyStr, "key2=value2")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	payload := map[string]string{"key1": "value1", "key2": "value2"}
	data, _ := json.Marshal(payload)

	resp, _, err := client.Post(ctx, server.URL, data, HeaderFormURLEncoded)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClient_Patch_Success(t *testing.T) {
	t.Parallel()

	expectedBody := map[string]string{"name": "updated"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, expectedBody, body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"updated"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	resp, body, err := client.Patch(ctx, server.URL, expectedBody)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"status":"updated"}`, string(body))
}

func TestHTTPClient_Put_Success(t *testing.T) {
	t.Parallel()

	expectedBody := map[string]string{"name": "replaced"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, expectedBody, body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"replaced"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	resp, body, err := client.Put(ctx, server.URL, expectedBody)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"status":"replaced"}`, string(body))
}

func TestHTTPClient_Delete_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	ctx := context.Background()
	resp, _, err := client.Delete(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHTTPClient_CustomResponseHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Original", "value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	responseHandlerCalled := false
	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ResponseHandler: func(rsp *http.Response) (*http.Response, error) {
			responseHandlerCalled = true
			rsp.Header.Set("X-Modified", "modified")
			return rsp, nil
		},
	}

	ctx := context.Background()
	resp, _, err := client.Get(ctx, server.URL)

	require.NoError(t, err)
	assert.True(t, responseHandlerCalled)
	assert.Equal(t, "value", resp.Header.Get("X-Original"))
	assert.Equal(t, "modified", resp.Header.Get("X-Modified"))
}

func TestHTTPClient_CustomShouldHandleError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	shouldHandleErrorCalled := false
	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ShouldHandleError: func(response *http.Response) bool {
			shouldHandleErrorCalled = true
			// Treat all responses as errors for testing
			return true
		},
		ErrorHandler: func(rsp *http.Response, body []byte) error {
			// Return nil to simulate ignoring the error
			return nil
		},
	}

	ctx := context.Background()
	resp, _, err := client.Get(ctx, server.URL)

	require.NoError(t, err)
	assert.True(t, shouldHandleErrorCalled)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClient_GetURL_AbsoluteURL(t *testing.T) {
	t.Parallel()

	client := &HTTPClient{}
	urlStr := "https://api.example.com/users"

	result, err := client.getURL(urlStr)

	require.NoError(t, err)
	assert.Equal(t, urlStr, result)
}

func TestHTTPClient_GetURL_RelativeURLWithBase(t *testing.T) {
	t.Parallel()

	client := &HTTPClient{
		Base: "https://api.example.com",
	}
	urlStr := "/users"

	result, err := client.getURL(urlStr)

	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/users", result)
}

func TestHTTPClient_GetURL_RelativeURLNoBase(t *testing.T) {
	t.Parallel()

	client := &HTTPClient{}
	urlStr := "/users"

	_, err := client.getURL(urlStr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty base URL")
}

func TestHTTPClient_SendRequest_NetworkError(t *testing.T) {
	t.Parallel()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	_, _, err = client.sendRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestHTTPClient_SendRequest_ResponseHandlerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("response handler error")
	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("test")),
				}, nil
			},
		},
		ResponseHandler: func(rsp *http.Response) (*http.Response, error) {
			return nil, expectedErr
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	_, _, err = client.sendRequest(req)

	require.ErrorIs(t, err, expectedErr)
}

func TestHTTPClient_SendRequest_ReadBodyError(t *testing.T) {
	t.Parallel()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Create a response with a body that will error when read
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(errorReader{}),
				}, nil
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	_, _, err = client.sendRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error reading response body")
}

func TestBodyReader_FormURLEncoded(t *testing.T) {
	t.Parallel()

	headers := Headers{HeaderFormURLEncoded}
	data := []byte(`{"key1":"value1","key2":"value2"}`)

	reader, length, err := bodyReader(headers, data)

	require.NoError(t, err)
	assert.Equal(t, int64(len("key1=value1&key2=value2")), length)

	readData, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Contains(t, string(readData), "key1=value1")
	assert.Contains(t, string(readData), "key2=value2")
}

func TestBodyReader_NonFormURLEncoded(t *testing.T) {
	t.Parallel()

	headers := Headers{{Key: "Content-Type", Value: "application/json"}}
	data := []byte(`{"test":"value"}`)

	reader, length, err := bodyReader(headers, data)

	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), length)

	readData, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, data, readData)
}

// ============================================================================
// REQUEST/RESPONSE BUILDING TESTS
// ============================================================================

func TestMakeGetRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{
		{Key: "Accept", Value: "application/json"},
		{Key: "X-Request-ID", Value: "123"},
	}

	req, err := MakeGetRequest(ctx, url, headers)

	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "123", req.Header.Get("X-Request-ID"))
}

func TestMakePostRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{
		{Key: "X-Custom", Value: "value"},
	}
	data := []byte(`{"test":"value"}`)

	req, err := makePostRequest(ctx, url, headers, data)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, data, body)
}

func TestMakePatchRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{
		{Key: "X-Custom", Value: "value"},
	}
	body := map[string]string{"test": "value"}

	req, err := makePatchRequest(ctx, url, headers, body)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestMakePatchRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{}

	// Create a body that cannot be marshaled to JSON
	body := make(chan int)

	_, err := makePatchRequest(ctx, url, headers, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request body is not valid JSON")
}

func TestMakePutRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{
		{Key: "X-Custom", Value: "value"},
	}
	body := map[string]string{"test": "value"}

	req, err := makePutRequest(ctx, url, headers, body)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestMakeDeleteRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	url := "https://example.com"
	headers := []Header{
		{Key: "X-Custom", Value: "value"},
	}

	req, err := makeDeleteRequest(ctx, url, headers)

	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
}

func TestAddHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	headers := []Header{
		{Key: "Accept", Value: "application/json"},
		{Key: "X-Request-ID", Value: "123"},
	}

	result := addHeaders(req, headers)

	assert.Equal(t, req, result)
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "123", req.Header.Get("X-Request-ID"))
}

func TestAddJSONContentTypeIfNotPresent_WhenAbsent(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)

	result := AddJSONContentTypeIfNotPresent(req)

	assert.Equal(t, req, result)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAddJSONContentTypeIfNotPresent_WhenPresent(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	result := AddJSONContentTypeIfNotPresent(req)

	assert.Equal(t, req, result)
	assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))
}

// ============================================================================
// URL HANDLING TESTS
// ============================================================================

func TestGetURL_Absolute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		baseURL  string
		url      string
		expected string
		err      bool
	}{
		{
			name:     "absolute URL with http",
			baseURL:  "",
			url:      "http://example.com/path",
			expected: "http://example.com/path",
			err:      false,
		},
		{
			name:     "absolute URL with https",
			baseURL:  "",
			url:      "https://example.com/path",
			expected: "https://example.com/path",
			err:      false,
		},
		{
			name:     "relative URL with base",
			baseURL:  "https://api.example.com",
			url:      "/users",
			expected: "https://api.example.com/users",
			err:      false,
		},
		{
			name:     "relative URL with base and path",
			baseURL:  "https://api.example.com/api",
			url:      "users",
			expected: "https://api.example.com/api/users",
			err:      false,
		},
		{
			name:     "relative URL without base",
			baseURL:  "",
			url:      "/users",
			expected: "",
			err:      true,
		},
		{
			name:     "empty base with empty URL",
			baseURL:  "",
			url:      "",
			expected: "",
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := getURL(tt.baseURL, tt.url)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestGetResponseBodyOnce(t *testing.T) {
	t.Parallel()

	// Test with valid response body
	bodyContent := `{"status":"ok"}`
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader(bodyContent)),
	}

	result := GetResponseBodyOnce(response)

	assert.Equal(t, []byte(bodyContent), result)
	// Don't assert anything about reading from the closed body
	// as the behavior is implementation-dependent
}

func TestGetResponseBodyOnce_ErrorReading(t *testing.T) {
	t.Parallel()

	// Create a response with a body that will error when read
	response := &http.Response{
		Body: io.NopCloser(errorReader{}),
	}

	result := GetResponseBodyOnce(response)

	assert.Nil(t, result)
}

// This test also that GetRequestHeaders preserves the canonicalized header names as they exist in the request.
func TestGetRequestHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  *http.Request
		expected Headers
	}{
		{
			name:     "nil request",
			request:  nil,
			expected: nil,
		},
		{
			name: "request with multiple headers",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Accept", "text/plain")
				req.Header.Add("X-Request-ID", "123")
				return req
			}(),
			expected: Headers{
				{Key: "Accept", Value: "application/json"},
				{Key: "Accept", Value: "text/plain"},
				{Key: "X-Request-Id", Value: "123"}, // Note: canonicalized form
			},
		},
		{
			name: "request with no headers",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
				return req
			}(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetRequestHeaders(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetResponseHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response *http.Response
		expected Headers
	}{
		{
			name:     "nil response",
			response: nil,
			expected: nil,
		},
		{
			name: "response with multiple headers",
			response: func() *http.Response {
				rsp := &http.Response{
					Header: make(http.Header),
				}
				rsp.Header.Add("Content-Type", "application/json")
				rsp.Header.Add("X-Request-ID", "123")
				rsp.Header.Add("Set-Cookie", "session=abc")
				rsp.Header.Add("Set-Cookie", "token=xyz")
				return rsp
			}(),
			expected: Headers{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "X-Request-Id", Value: "123"},
				{Key: "Set-Cookie", Value: "session=abc"},
				{Key: "Set-Cookie", Value: "token=xyz"},
			},
		},
		{
			name: "response with no headers",
			response: &http.Response{
				Header: make(http.Header),
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetResponseHeaders(tt.response)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// CONTEXT HANDLING TESTS
// ============================================================================

func TestHTTPClient_MethodsWithContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context is passed through
		assert.NotNil(t, r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
	}

	var tk testKey = "test-key"

	ctx := context.WithValue(context.Background(), tk, "test-value")

	// Test all HTTP methods with context
	_, _, err := client.Get(ctx, server.URL)
	require.NoError(t, err)

	_, _, err = client.Post(ctx, server.URL, []byte(`{"test":"value"}`))
	require.NoError(t, err)

	_, _, err = client.Patch(ctx, server.URL, map[string]string{"test": "value"})
	require.NoError(t, err)

	_, _, err = client.Put(ctx, server.URL, map[string]string{"test": "value"})
	require.NoError(t, err)

	_, _, err = client.Delete(ctx, server.URL)
	require.NoError(t, err)
}

func TestHTTPClient_WithEmptyResponseHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ResponseHandler: nil, // Explicitly nil
	}

	ctx := context.Background()
	resp, _, err := client.Get(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClient_WithEmptyErrorHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ErrorHandler: nil, // Explicitly nil, should use default error handler
	}

	ctx := context.Background()
	_, _, err := client.Get(ctx, server.URL)

	require.Error(t, err)
	// Should use InterpretError from the default error handling
}

func TestHTTPClient_WithEmptyShouldHandleError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &HTTPClient{
		Client: &mockAuthClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			},
		},
		ShouldHandleError: nil, // Should use default 2xx check
	}

	ctx := context.Background()
	resp, _, err := client.Get(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHeaderModeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, headerModeUnset)
	assert.Equal(t, 1, HeaderModeAppend)
	assert.Equal(t, 2, HeaderModeOverwrite)
	assert.Equal(t, 3, HeaderModeSetIfMissing)
}

func TestHeaderFormURLEncodedConstant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Content-Type", HeaderFormURLEncoded.Key)
	assert.Equal(t, "application/x-www-form-urlencoded", HeaderFormURLEncoded.Value)
}
