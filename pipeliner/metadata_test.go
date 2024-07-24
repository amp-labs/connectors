package pipeliner

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []struct {
		name         string
		input        []string
		server       *httptest.Server
		expected     *common.ListObjectMetadataResult
		expectedErrs []error
	}{
		{
			name:  "At least one object name must be queried",
			input: nil,
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:  "Unknown object requested",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{scrapper.ErrObjectNotFound},
		},
		{
			name:  "Successfully describe one object with metadata",
			input: []string{"Notes"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Notes": {
						DisplayName: "Notes",
						FieldsMap: map[string]string{
							"account":             "account",
							"account_id":          "account_id",
							"contact":             "contact",
							"contact_id":          "contact_id",
							"created":             "created",
							"custom_entity":       "custom_entity",
							"custom_entity_id":    "custom_entity_id",
							"id":                  "id",
							"is_delete_protected": "is_delete_protected",
							"is_deleted":          "is_deleted",
							"lead":                "lead",
							"lead_oppty_id":       "lead_oppty_id",
							"modified":            "modified",
							"note":                "note",
							"oppty":               "oppty",
							"owner":               "owner",
							"owner_id":            "owner_id",
							"project":             "project",
							"project_id":          "project_id",
							"quote":               "quote",
							"quote_id":            "quote_id",
							"revision":            "revision",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
		{
			name:  "Successfully describe multiple objects with metadata",
			input: []string{"Phones", "Tags"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Phones": {
						DisplayName: "Phones",
						FieldsMap: map[string]string{
							"call_forwarding_phone":    "call_forwarding_phone",
							"created":                  "created",
							"id":                       "id",
							"is_delete_protected":      "is_delete_protected",
							"is_deleted":               "is_deleted",
							"message_forwarding_email": "message_forwarding_email",
							"modified":                 "modified",
							"name":                     "name",
							"owner":                    "owner",
							"owner_id":                 "owner_id",
							"phone_number":             "phone_number",
							"revision":                 "revision",
							"twilio_caller_sid":        "twilio_caller_sid",
							"type":                     "type",
						},
					},
					"Tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"color":               "color",
							"created":             "created",
							"creator":             "creator",
							"creator_id":          "creator_id",
							"id":                  "id",
							"is_delete_protected": "is_delete_protected",
							"is_deleted":          "is_deleted",
							"modified":            "modified",
							"name":                "name",
							"revision":            "revision",
							"supported_entities":  "supported_entities",
							"use_lang":            "use_lang",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.ListObjectMetadata(context.Background(), tt.input)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			// check every error
			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)",
					tt.name, tt.expected, output, diff)
			}
		})
	}
}
