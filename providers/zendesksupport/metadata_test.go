package zendesksupport

import (
	"context"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadataZendeskSupportModule(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseTicketsCustomFields := testutils.DataFromFile(t, "read/custom_fields/ticket_fields.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Object coming from different module is unknown",
			Input:        []string{"articles"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"brands"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
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
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"bookmarks", "ticket_audits"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
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
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe ticket custom fields",
			Input: []string{"tickets"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/api/v2/ticket_fields"),
				Then:  mockserver.Response(http.StatusOK, responseTicketsCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tickets": {
						DisplayName: "Tickets",
						Fields: map[string]common.FieldMetadata{
							"comment": {
								DisplayName:  "comment",
								ValueType:    "other",
								ProviderType: "object",
								ReadOnly:     false,
							},
							"priority": {
								DisplayName:  "priority",
								ValueType:    "singleSelect",
								ProviderType: "string",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "urgent",
									DisplayValue: "urgent",
								}, {
									Value:        "high",
									DisplayValue: "high",
								}, {
									Value:        "normal",
									DisplayValue: "normal",
								}, {
									Value:        "low",
									DisplayValue: "low",
								}},
							},
							// Custom field
							"Customer Type": {
								DisplayName:  "Customer Type",
								ValueType:    "singleSelect",
								ProviderType: "tagger",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "vip_customer",
									DisplayValue: "VIP Customer",
								}, {
									Value:        "standard_customer",
									DisplayValue: "Standard Customer",
								}},
							},
						},
						FieldsMap: map[string]string{
							"comment": "comment",
							// Custom fields
							"Customer Type": "Customer Type",
							"Topic":         "Topic",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleTicketing)
			})
		})
	}
}

func TestListObjectMetadataHelpCenterModule(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "Object coming from different module is unknown",
			Input:        []string{"brands"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"articles"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"articles": {
						DisplayName: "Articles",
						FieldsMap: map[string]string{
							"author_id":           "author_id",
							"body":                "body",
							"comments_disabled":   "comments_disabled",
							"content_tag_ids":     "content_tag_ids",
							"created_at":          "created_at",
							"draft":               "draft",
							"edited_at":           "edited_at",
							"html_url":            "html_url",
							"id":                  "id",
							"label_names":         "label_names",
							"locale":              "locale",
							"outdated":            "outdated",
							"outdated_locales":    "outdated_locales",
							"permission_group_id": "permission_group_id",
							"position":            "position",
							"promoted":            "promoted",
							"section_id":          "section_id",
							"source_locale":       "source_locale",
							"title":               "title",
							"updated_at":          "updated_at",
							"url":                 "url",
							"user_segment_id":     "user_segment_id",
							"user_segment_ids":    "user_segment_ids",
							"vote_count":          "vote_count",
							"vote_sum":            "vote_sum",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleHelpCenter)
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
