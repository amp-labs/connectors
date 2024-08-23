package zendesksupport

import (
	"context"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{scrapper.ErrObjectNotFound},
		},
		{
			Name:   "Successfully describe one object with metadata",
			Input:  []string{"brands"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"brands": {
						DisplayName: "Brands",
						FieldsMap: map[string]string{
							"active":             "active",
							"brand_url":          "brand_url",
							"created_at":         "created_at",
							"default":            "default",
							"has_help_center":    "has_help_center",
							"help_center_state":  "help_center_state",
							"host_mapping":       "host_mapping",
							"id":                 "id",
							"is_deleted":         "is_deleted",
							"logo":               "logo",
							"name":               "name",
							"signature_template": "signature_template",
							"subdomain":          "subdomain",
							"ticket_form_ids":    "ticket_form_ids",
							"updated_at":         "updated_at",
							"url":                "url",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"bookmarks", "ticket_audits"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"bookmarks": {
						DisplayName: "Bookmarks",
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"id":         "id",
							"ticket":     "ticket",
							"url":        "url",
						},
					},
					"ticket_audits": {
						DisplayName: "Ticket Audits",
						FieldsMap: map[string]string{
							"author_id":  "author_id",
							"created_at": "created_at",
							"events":     "events",
							"id":         "id",
							"metadata":   "metadata",
							"ticket_id":  "ticket_id",
							"via":        "via",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func BenchmarkListObjectMetadata(b *testing.B) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		b.Fatalf("%s: couldn't initialize connector", err)
	}

	dummyServer := mockserver.Dummy()

	connector.setBaseURL(dummyServer.URL)

	// start of benchmark
	for i := 0; i < b.N; i++ {
		_, _ = connector.ListObjectMetadata(context.Background(), []string{"brands"})
	}
}
