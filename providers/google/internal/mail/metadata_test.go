package mail

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name: "Successful metadata for messages and drafts",
			Input: []string{
				"messages", "drafts",
				"AMPERSAND-messages", "AMPERSAND-sentMessages",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"messages": {
						DisplayName: "Messages",
						Fields: map[string]common.FieldMetadata{
							"threadId": {
								DisplayName:  "Thread Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     nil,
							},
						},
					},
					"drafts": {
						DisplayName: "Drafts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
							},
							"message": {
								DisplayName:  "Message",
								ValueType:    "other",
								ProviderType: "object",
								ReadOnly:     new(false),
							},
						},
					},
					"AMPERSAND-sentMessages": {
						DisplayName: "Sent Messages",
						Fields: map[string]common.FieldMetadata{
							"threadId": {
								DisplayName:  "Thread Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     nil,
							},
						},
					},
					"AMPERSAND-messages": {
						DisplayName: "Messages",
						Fields: map[string]common.FieldMetadata{
							"threadId": {
								DisplayName:  "Thread Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     nil,
							},
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestAdapter(tt.Server)
			})
		})
	}
}

func constructTestAdapter(server *httptest.Server) (*Adapter, error) {
	adapter, err := NewAdapter(
		common.ConnectorParams{
			Module:              providers.ModuleGoogleGmail,
			AuthenticatedClient: server.Client(),
		},
	)
	if err != nil {
		return nil, err
	}

	adapter.SetUnitTestMockServerBaseURL(server.URL)

	return adapter, nil
}
