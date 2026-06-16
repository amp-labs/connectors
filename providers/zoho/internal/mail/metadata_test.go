package mail

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	accountsResponse := testutils.DataFromFile(t, "accounts.json")

	t.Run("at least one object name must be provided", func(t *testing.T) {
		t.Parallel()

		adapter := constructTestAdapter(t, mockserver.Dummy().URL)

		_, err := adapter.ListObjectMetadata(context.Background(), nil)
		if !errors.Is(err, common.ErrMissingObjects) {
			t.Fatalf("expected ErrMissingObjects, got %v", err)
		}
	})

	t.Run("unsupported object collects per-object error", func(t *testing.T) {
		t.Parallel()

		adapter := constructTestAdapter(t, mockserver.Dummy().URL)

		result, err := adapter.ListObjectMetadata(context.Background(), []string{"folders"})
		if err != nil {
			t.Fatalf("unexpected top-level error: %v", err)
		}

		if !errors.Is(result.Errors["folders"], common.ErrObjectNotSupported) {
			t.Fatalf("expected ErrObjectNotSupported for folders, got %v", result.Errors["folders"])
		}
	})

	t.Run("samples fields from accounts endpoint", func(t *testing.T) {
		t.Parallel()

		server := mockserver.Switch{
			Setup: mockserver.ContentJSON(),
			Cases: []mockserver.Case{{
				If:   mockcond.Path("/api/accounts"),
				Then: mockserver.Response(http.StatusOK, accountsResponse),
			}},
		}.Server()
		defer server.Close()

		adapter := constructTestAdapter(t, server.URL)

		result, err := adapter.ListObjectMetadata(context.Background(), []string{"accounts"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if e := result.Errors["accounts"]; e != nil {
			t.Fatalf("unexpected per-object error: %v", e)
		}

		fields := result.Result["accounts"].Fields

		assertFieldType(t, fields, "accountId", common.ValueTypeString)
		assertFieldType(t, fields, "primaryEmailAddress", common.ValueTypeString)
		assertFieldType(t, fields, "incomingBlocked", common.ValueTypeBoolean)
		assertFieldType(t, fields, "sendMailDetails", common.ValueTypeOther)
	})
}

// TestParseMetadataResponse_RecordPaths verifies records are located regardless of
// which key (or nested key) holds the array across Zoho Mail endpoints.
func TestParseMetadataResponse_RecordPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		obj  objectDescriptor
		body string
	}{
		{
			name: "top-level key",
			obj:  objectDescriptor{recordsPath: []string{"data"}},
			body: `{"status":{"code":200},"data":[{"id":"1","done":false}]}`,
		},
		{
			name: "different top-level key",
			obj:  objectDescriptor{recordsPath: []string{"list"}},
			body: `{"status":{"code":200},"list":[{"id":"1","done":false}]}`,
		},
		{
			name: "nested key",
			obj:  objectDescriptor{recordsPath: []string{"data", "lists"}},
			body: `{"status":{"code":200},"data":{"lists":[{"id":"1","done":false}]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metadata, err := parseMetadataResponse("obj", tt.obj, newJSONResponse(t, tt.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertFieldType(t, metadata.Fields, "id", common.ValueTypeString)
			assertFieldType(t, metadata.Fields, "done", common.ValueTypeBoolean)
		})
	}
}

func newJSONResponse(t *testing.T, body string) *common.JSONHTTPResponse {
	t.Helper()

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	resp, err := common.ParseJSONResponse(context.Background(), httpResp, []byte(body))
	if err != nil {
		t.Fatalf("ParseJSONResponse: %v", err)
	}

	return resp
}

func assertFieldType(t *testing.T, fields common.FieldsMetadata, name string, want common.ValueType) {
	t.Helper()

	field, ok := fields[name]
	if !ok {
		t.Fatalf("expected field %q to be present", name)
	}

	if field.ValueType != want {
		t.Fatalf("field %q: got value type %q, want %q", name, field.ValueType, want)
	}
}

func constructTestAdapter(t *testing.T, baseURL string) *Adapter {
	t.Helper()

	client := &common.JSONHTTPClient{
		HTTPClient: &common.HTTPClient{
			Client: mockutils.NewClient(),
		},
	}

	adapter, err := NewAdapter(client, &providers.ModuleInfo{BaseURL: baseURL}, "")
	if err != nil {
		t.Fatalf("failed to construct adapter: %v", err)
	}

	return adapter
}
