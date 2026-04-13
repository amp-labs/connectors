package calendly

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/stretchr/testify/require"
)

func TestProjectFields(t *testing.T) {
	t.Parallel()

	resource := map[string]any{
		"Name":   "One-on-One",
		"uri":    "https://api.calendly.com/event_types/abc",
		"active": true,
	}

	t.Run("empty field set returns all keys lowercased", func(t *testing.T) {
		t.Parallel()

		got := projectFields(resource, datautils.NewSetFromList([]string{}))
		require.Equal(t, map[string]any{
			"name":   "One-on-One",
			"uri":    "https://api.calendly.com/event_types/abc",
			"active": true,
		}, got)
	})

	t.Run("filters to requested fields", func(t *testing.T) {
		t.Parallel()

		fs := datautils.NewSetFromList([]string{"name", "active"})
		got := projectFields(resource, fs)
		require.Equal(t, map[string]any{
			"name":   "One-on-One",
			"active": true,
		}, got)
	})
}

func TestLowercaseKeysCopy(t *testing.T) {
	t.Parallel()

	in := map[string]any{"Foo": 1, "BAR": "x"}
	got := lowercaseKeysCopy(in)
	require.Equal(t, map[string]any{"foo": 1, "bar": "x"}, got)
}

func TestGetRecordsByIds_errors(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: http.DefaultClient,
	})
	require.NoError(t, err)
	ctx := t.Context()

	_, err = conn.GetRecordsByIds(ctx, "contacts", []string{"x"}, nil, nil)
	require.ErrorIs(t, err, common.ErrNotImplemented)

	_, err = conn.GetRecordsByIds(ctx, "event_types", nil, nil, nil)
	require.ErrorIs(t, err, common.ErrMissingRecordID)
}

func TestGetRecordsByIds_success(t *testing.T) {
	t.Parallel()

	const path = "/event_types/11111111-1111-1111-1111-111111111111"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, path, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		fullURI := "http://" + r.Host + path
		body, err := json.Marshal(map[string]any{
			"resource": map[string]any{
				"uri":    fullURI,
				"name":   "Coffee Chat",
				"active": true,
			},
		})
		require.NoError(t, err)
		_, err = w.Write(body)
		require.NoError(t, err)
	}))
	t.Cleanup(srv.Close)

	recordURI := srv.URL + path

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: http.DefaultClient,
	})
	require.NoError(t, err)

	rows, err := conn.GetRecordsByIds(t.Context(), "event_types", []string{recordURI},
		[]string{"name", "active"}, nil)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	require.Equal(t, recordURI, rows[0].Id)
	require.Equal(t, map[string]any{
		"name":   "Coffee Chat",
		"active": true,
	}, rows[0].Fields)
	require.Equal(t, "Coffee Chat", rows[0].Raw["name"])
}

func TestGetRecordsByIds_skipsEmptyURI(t *testing.T) {
	t.Parallel()

	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"resource":{"uri":"u","name":"only"}}`))
		require.NoError(t, err)
	}))
	t.Cleanup(srv.Close)

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: http.DefaultClient,
	})
	require.NoError(t, err)

	rows, err := conn.GetRecordsByIds(t.Context(), "event_types",
		[]string{"", srv.URL + "/event_types/one"}, nil, nil)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, 1, callCount)
}

func TestGetRecordsByIds_emptyResource(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"resource":null}`))
		require.NoError(t, err)
	}))
	t.Cleanup(srv.Close)

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: http.DefaultClient,
	})
	require.NoError(t, err)

	_, err = conn.GetRecordsByIds(t.Context(), "event_types", []string{srv.URL + "/x"}, nil, nil)
	require.ErrorContains(t, err, "empty resource")
}
