package mailgun

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object returns ErrObjectNotSupported",
			Input:      []string{"nonexistent"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent": common.ErrObjectNotSupported,
				},
			},
		},
		{
			// Metadata is served statically from the embedded OpenAPI schema,
			// so no provider call is made (Dummy server).
			Name:       "Describe account-scoped domains object",
			Input:      []string{"domains"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"domains": {
						DisplayName: "Domains",
						FieldsMap: map[string]string{
							"id":         "id",
							"name":       "name",
							"state":      "state",
							"created_at": "created_at",
						},
					},
				},
			},
		},
		{
			// Slash-named sub-resource resolves correctly.
			Name:       "Describe slash-named lists/members object",
			Input:      []string{"lists/members"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists/members": {
						DisplayName: "Mailing List Members",
						FieldsMap: map[string]string{
							"address":    "address",
							"subscribed": "subscribed",
						},
					},
				},
			},
		},
		{
			// The analytics/logs object comes from the POST /v1/analytics/logs
			// endpoint, proving POST-sourced schemas are included in metadata.
			Name:       "Describe POST-sourced analytics/logs object",
			Input:      []string{"analytics/logs"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"analytics/logs": {
						DisplayName: "Logs",
						FieldsMap: map[string]string{
							"id":        "id",
							"event":     "event",
							"recipient": "recipient",
						},
					},
				},
			},
		},
		{
			Name:       "Describe multiple objects at once",
			Input:      []string{"webhooks", "routes", "tags"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"webhooks": {
						DisplayName: "Webhooks",
						FieldsMap: map[string]string{
							"url":        "url",
							"webhook_id": "webhook_id",
						},
					},
					"routes": {
						DisplayName: "Routes",
						FieldsMap: map[string]string{
							"id":         "id",
							"expression": "expression",
						},
					},
					"tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"tag": "tag",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
