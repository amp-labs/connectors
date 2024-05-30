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
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	fakeServerResp := mockutils.DataFromFile(t, "metadata.xml")

	tests := []struct {
		name                string
		input               []string
		server              *httptest.Server
		connector           Connector
		expected            *common.ListObjectMetadataResult
		expectedFieldsCount map[string]int // used instead of `expected` when response result is too big
		expectedErrs        []error
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
			name:  "Missing XML root",
			input: []string{"msfp_surveyinvite"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `<?xml version="1.0" encoding="utf-8"?>`)
			})),
			expectedErrs: []error{common.ErrNoXMLRoot},
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
			input: []string{"accountleads", "adx_invitation_invitecontacts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accountleads": {
						DisplayName: "accountleads",
						FieldsMap: map[string]string{
							"accountid":                 "accountid",
							"accountleadid":             "accountleadid",
							"importsequencenumber":      "importsequencenumber",
							"leadid":                    "leadid",
							"name":                      "name",
							"overriddencreatedon":       "overriddencreatedon",
							"timezoneruleversionnumber": "timezoneruleversionnumber",
							"utcconversiontimezonecode": "utcconversiontimezonecode",
							"versionnumber":             "versionnumber",
						},
					},
					"adx_invitation_invitecontacts": {
						DisplayName: "adx_invitation_invitecontacts",
						FieldsMap: map[string]string{
							"adx_invitation_invitecontactsid": "adx_invitation_invitecontactsid",
							"adx_invitationid":                "adx_invitationid",
							"contactid":                       "contactid",
							"versionnumber":                   "versionnumber",
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
			input: []string{"phonecall", "activitypointer", "crmbaseentity"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			expectedFieldsCount: map[string]int{
				"phonecall":       65,
				"activitypointer": 58,
				"crmbaseentity":   0,
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
			// there are 2 modes we can compare, by exact field values or matching quantity
			if len(tt.expectedFieldsCount) != 0 {
				// we are comparing if number of fields match under ListObjectMetadataResult.Result
				for entityName, count := range tt.expectedFieldsCount {
					entity, ok := output.Result[entityName]
					if !ok {
						t.Fatalf("%s: expected entity was missing: (%v)", tt.name, entityName)
					}

					got := len(entity.FieldsMap)
					if got != count {
						t.Fatalf("%s: expected entity '%v' to have (%v) fields got: (%v)",
							tt.name, entityName, count, got)
					}
				}
			} else { // nolint:gocritic
				// usual comparison of ListObjectMetadataResult
				if !reflect.DeepEqual(output, tt.expected) {
					diff := deep.Equal(output, tt.expected)
					t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)",
						tt.name, tt.expected, output, diff)
				}
			}
		})
	}
}
