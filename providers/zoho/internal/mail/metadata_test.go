package mail

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	accountsResponse := testutils.DataFromFile(t, "accounts.json")
	notesResponse := testutils.DataFromFile(t, "notes.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object collects per-object error",
			Input:      []string{"folders"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"folders": common.ErrObjectNotSupported,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Samples fields from a static endpoint",
			Input: []string{"accounts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/accounts"),
					Then: mockserver.Response(http.StatusOK, accountsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accounts": {
						DisplayName: "accounts",
						Fields: map[string]common.FieldMetadata{
							"accountId":           {DisplayName: "accountId", ValueType: common.ValueTypeString},
							"primaryEmailAddress": {DisplayName: "primaryEmailAddress", ValueType: common.ValueTypeString},
							"incomingBlocked":     {DisplayName: "incomingBlocked", ValueType: common.ValueTypeBoolean},
							"sendMailDetails":     {DisplayName: "sendMailDetails", ValueType: common.ValueTypeOther},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Samples fields from a nested, paginated endpoint",
			Input: []string{"notes"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					// notes supports pagination, so the sampler must send limit=1.
					If: mockcond.And{
						mockcond.Path("/api/notes/me"),
						mockcond.QueryParam("limit", "1"),
					},
					Then: mockserver.Response(http.StatusOK, notesResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"notes": {
						DisplayName: "notes",
						Fields: map[string]common.FieldMetadata{
							"entityId":   {DisplayName: "entityId", ValueType: common.ValueTypeString},
							"title":      {DisplayName: "title", ValueType: common.ValueTypeString},
							"isFavorite": {DisplayName: "isFavorite", ValueType: common.ValueTypeBoolean},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tc := testroutines.TestCase[[]string, *common.ListObjectMetadataResult](tt)
			t.Cleanup(tc.Close)

			adapter := constructTestAdapter(t, tt.Server.URL, "")

			output, err := adapter.ListObjectMetadata(t.Context(), tc.Input)
			tc.Validate(t, err, output)
		})
	}
}

func constructTestAdapter(t *testing.T, baseURL, accountID string) *Adapter {
	t.Helper()

	client := &common.JSONHTTPClient{
		HTTPClient: &common.HTTPClient{
			Client: mockutils.NewClient(),
		},
	}

	adapter, err := NewAdapter(client, &providers.ModuleInfo{BaseURL: baseURL}, accountID)
	if err != nil {
		t.Fatalf("failed to construct adapter: %v", err)
	}

	return adapter
}
