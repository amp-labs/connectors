// nolint:revive
package common

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type dummyTransport struct{}

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
