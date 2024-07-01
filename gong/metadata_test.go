package gong

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
			input: []string{"flows"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"flows": {
						DisplayName: "List Gong Engage flows (/v2/flows)",
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"folderId":     "folderId",
							"folderName":   "folderName",
							"visibility":   "visibility",
							"creationDate": "creationDate",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
		{
			name:  "Successfully describe multiple objects with metadata",
			input: []string{"workspaces", "logs"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workspaces": {
						DisplayName: "List all company workspaces (/v2/workspaces)",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"description": "description",
						},
					},
					"logs": {
						DisplayName: "Retrieve logs data by type and time range (/v2/logs)",
						FieldsMap: map[string]string{
							"userId":                   "userId",
							"userEmailAddress":         "userEmailAddress",
							"userFullName":             "userFullName",
							"impersonatorUserId":       "impersonatorUserId",
							"impersonatorEmailAddress": "impersonatorEmailAddress",
							"impersonatorFullName":     "impersonatorFullName",
							"impersonatorCompanyId":    "impersonatorCompanyId",
							"eventTime":                "eventTime",
							"logRecord":                "logRecord",
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
