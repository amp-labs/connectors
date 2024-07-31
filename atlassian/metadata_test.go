package atlassian

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseIssueSchema := testutils.DataFromFile(t, "issue-metadata.json")

	tests := []struct {
		name         string
		input        []string
		server       *httptest.Server
		comparator   func(serverURL string, actual, expected *common.ListObjectMetadataResult) bool
		expected     *common.ListObjectMetadataResult
		expectedErrs []error
	}{
		{
			name:         "Mime response header expected",
			input:        []string{"butterflies"},
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Server response must include array",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, ``)
			})),
			expectedErrs: []error{ErrParsingMetadata},
		},
		{
			name:  "Server response must have at least one field",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `[]`)
			})),
			expectedErrs: []error{
				ErrMissingMetadata,
				ErrParsingMetadata,
			},
		},
		{
			name:  "Field response must have identifier",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `[{}]`)
			})),
			expectedErrs: []error{ErrParsingMetadata},
		},
		{
			name:  "Field response must have display name",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `[{"id": "issuerestriction"}]`)
			})),
			expectedErrs: []error{ErrParsingMetadata},
		},
		{
			name:  "Successfully describe Issue metadata",
			input: []string{},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseIssueSchema)
			})),
			comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"issue": {
						DisplayName: "Issue",
						FieldsMap: map[string]string{
							// Manually attached fields:
							"id": "Identifier",
							// Fields coming from server response:
							"issuekey":                      "Key",
							"priority":                      "Priority",
							"creator":                       "Creator",
							"worklog":                       "Log Work",
							"issuetype":                     "Issue Type",
							"issuelinks":                    "Linked Issues",
							"fixVersions":                   "Fix versions",
							"issuerestriction":              "Restrict to",
							"statuscategorychangedate":      "Status Category Changed",
							"aggregatetimeoriginalestimate": "Σ Original Estimate",
							"customfield_10028":             "Submitted forms",
							"customfield_10035":             "Project overview key",
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
				WithModule(ModuleJira),
				WithMetadata(map[string]string{
					"cloudID": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // it doesn't matter for the test
				}),
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

			// compare desired output
			var ok bool
			if tt.comparator == nil {
				// default comparison is concerned about all fields
				ok = reflect.DeepEqual(output, tt.expected)
			} else {
				ok = tt.comparator(tt.server.URL, output, tt.expected)
			}

			if !ok {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
