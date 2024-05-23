package intercom

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scrapper"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCompanyAttributes := mockutils.DataFromFile(t, "metadata-attributes-company.json")
	responseUnsupportedModel := mockutils.DataFromFile(t, "unsupported-model.json")

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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseUnsupportedModel)
			})),
			expectedErrs: []error{scrapper.ErrObjectNotFound},
		},
		{
			// Some metadata is coming from ListDataAttributes and some from static file.
			name:  "Successfully describe one object with hybrid metadata",
			input: []string{"company"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseCompanyAttributes)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "Companies",
						FieldsMap: map[string]string{
							// Fields from static file
							"custom_attributes": "custom_attributes",
							"plan":              "plan",
							"segments":          "segments",
							"tags":              "tags",
							"type":              "type",
							// Fields from API endpoint (all present in static but different display)
							"my_custom_field":   "Custom field for test purposes",
							"last_request_at":   "Company last seen",
							"name":              "Plan",
							"company_id":        "Company ID",
							"updated_at":        "Updated at",
							"monthly_spend":     "Monthly Spend",
							"size":              "Company size",
							"industry":          "Company industry",
							"website":           "Company website",
							"id":                "Plan ID",
							"created_at":        "Created at",
							"app_id":            "App ID",
							"remote_created_at": "Company created at",
							"session_count":     "Company web sessions",
							"user_count":        "People",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
		{
			name:  "Successfully describe multiple objects with metadata",
			input: []string{"data_events", "teams"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseUnsupportedModel)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"data_events": {
						DisplayName: "Data Events",
						FieldsMap: map[string]string{
							"created_at":       "created_at",
							"email":            "email",
							"event_name":       "event_name",
							"id":               "id",
							"intercom_user_id": "intercom_user_id",
							"metadata":         "metadata",
							"type":             "type",
							"user_id":          "user_id",
						},
					},
					"teams": {
						DisplayName: "Teams",
						FieldsMap: map[string]string{
							"admin_ids":            "admin_ids",
							"admin_priority_level": "admin_priority_level",
							"id":                   "id",
							"name":                 "name",
							"type":                 "type",
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
