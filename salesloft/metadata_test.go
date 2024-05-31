package salesloft

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
			input: []string{"bulk-jobs-results"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"bulk-jobs-results": {
						DisplayName: "List job data for a completed bulk job.",
						FieldsMap: map[string]string{
							"error":    "error",
							"id":       "id",
							"record":   "record",
							"resource": "resource",
							"status":   "status",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
		{
			name:  "Successfully describe multiple objects with metadata",
			input: []string{"account-tiers", "actions"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account-tiers": {
						DisplayName: "List Account Tiers",
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"id":         "id",
							"name":       "name",
							"order":      "order",
							"updated_at": "updated_at",
						},
					},
					"actions": {
						DisplayName: "List actions",
						FieldsMap: map[string]string{
							"action_details":      "action_details",
							"cadence":             "cadence",
							"created_at":          "created_at",
							"due":                 "due",
							"due_on":              "due_on",
							"id":                  "id",
							"multitouch_group_id": "multitouch_group_id",
							"person":              "person",
							"status":              "status",
							"step":                "step",
							"task":                "task",
							"type":                "type",
							"updated_at":          "updated_at",
							"user":                "user",
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
