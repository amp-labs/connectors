package common

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModifier(t *testing.T) {
	t.Parallel()

	modifier, ok := getRequestModifier(context.Background())
	require.False(t, ok)
	require.Nil(t, modifier)

	ctx := WithRequestModifier(context.Background(), func(req *http.Request) {
		req.Header.Set("Header", "value")
	})

	modifier, ok = getRequestModifier(ctx)
	require.True(t, ok)

	req, err := http.NewRequest("GET", "https://example.com", nil)
	require.NoError(t, err)

	modifier(req)

	require.Equal(t, "value", req.Header.Get("Header"))
}

func TestApplySetIfMissingEmptyHeadersRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "https://example.com", nil)
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

	req, err := http.NewRequest("GET", "https://example.com", nil)
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

	req, err := http.NewRequest("GET", "https://example.com", nil)
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

	req, err := http.NewRequest("GET", "https://example.com", nil)
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

	req, err := http.NewRequest("GET", "https://example.com", nil)
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
