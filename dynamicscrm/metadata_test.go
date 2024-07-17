package dynamicscrm

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
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	fakeServerResp := testutils.DataFromFile(t, "metadata.xml")

	tests := []struct {
		name         string
		input        []string
		server       *httptest.Server
		connector    Connector
		comparator   func(serverURL string, actual, expected *common.ListObjectMetadataResult) bool
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
			name:  "Mime response header expected",
			input: []string{"msfp_surveyinvite"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Missing XML response on status OK",
			input: []string{"msfp_surveyinvite"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, "")
			})),
			expectedErrs: []error{common.ErrNotXML},
		},
		{
			name:  "Server response without CRM Schema",
			input: []string{"msfp_surveyinvite"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				<?xml version="1.0" encoding="utf-8"?>
				<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"></edmx:Edmx>`)
			})),
			expectedErrs: []error{ErrMissingSchema},
		},
		{
			name:  "Object name cannot be found from server response",
			input: []string{"msfp_surveyinvite"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			expectedErrs: []error{ErrObjectNotFound, errors.New("unknown entity msfp_surveyinvite")}, // nolint:goerr113
		},
		{
			name:  "Correctly list metadata for account leads and invite contact",
			input: []string{"chats", "faxes"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"faxes": {
						DisplayName: "faxes",
						FieldsMap: map[string]string{
							"tsid":                 "tsid",
							"numberofpages":        "numberofpages",
							"coverpagename":        "coverpagename",
							"overriddencreatedon":  "overriddencreatedon",
							"subcategory":          "subcategory",
							"billingcode":          "billingcode",
							"subscriptionid":       "subscriptionid",
							"importsequencenumber": "importsequencenumber",
							"directioncode":        "directioncode",
							"faxnumber":            "faxnumber",
							"category":             "category",
						},
					},
					"chats": {
						DisplayName: "chats",
						FieldsMap: map[string]string{
							"modifiedinteamson":                "modifiedinteamson",
							"_linkedby_value":                  "_linkedby_value",
							"_unlinkedby_value":                "_unlinkedby_value",
							"teamschatid":                      "teamschatid",
							"eventssummary":                    "eventssummary",
							"importsequencenumber":             "importsequencenumber",
							"overriddencreatedon":              "overriddencreatedon",
							"linkedon":                         "linkedon",
							"unlinkedon":                       "unlinkedon",
							"lastsyncerror":                    "lastsyncerror",
							"modifiedinteamsbyactivitypartyid": "modifiedinteamsbyactivitypartyid",
							"syncstatus":                       "syncstatus",
							"formattedscheduledstart":          "formattedscheduledstart",
							"statuscode":                       "statuscode",
						},
					},
				},
				Errors: nil,
			},
			expectedErrs: nil,
		},
		{
			// In total phonecall will have 65 fields, where
			// phonecall 		(has 7 fields) and inherits from
			// activitypointer 	(has 58 fields), which in turn inherits from
			// crmbaseentity 	(has 0 fields)
			name:  "Correctly list metadata for phone calls including inherited fields",
			input: []string{"phonecalls", "activitypointers", "crmbaseentities"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				// we are comparing if number of fields match under ListObjectMetadataResult.Result
				return compareFieldCount(map[string]int{
					"phonecalls":       65,
					"activitypointers": 58,
					"crmbaseentities":  0,
				}, actual)
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

func compareFieldCount(expectedFieldsCount map[string]int, actual *common.ListObjectMetadataResult) bool {
	for entityName, count := range expectedFieldsCount {
		if entity, ok := actual.Result[entityName]; !ok {
			return false
		} else {
			got := len(entity.FieldsMap)
			if got != count {
				return false
			}
		}
	}

	return true
}
