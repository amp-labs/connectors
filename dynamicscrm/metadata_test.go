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
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactsSchema := testutils.DataFromFile(t, "contacts-schema.json")
	// Attributes file is a shorter form of real Microsoft server response.
	responseContactsAttributes := testutils.DataFromFile(t, "contacts-attributes.json")

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
			name:         "At least one object name must be queried",
			input:        nil,
			server:       mockserver.Dummy(),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:         "Mime response header expected",
			input:        []string{"accounts"},
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Schema endpoint is not available for object",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte{})
			})),
			expectedErrs: []error{ErrObjectNotFound},
		},
		{
			name:  "Attributes endpoint is not available for object",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "EntityDefinitions(LogicalName='butterfly')"):
					_, _ = w.Write(responseContactsSchema)
				default:
					_, _ = w.Write([]byte{})
				}
			})),
			expectedErrs: []error{ErrObjectNotFound},
		},
		{
			name:  "Object doesn't have attributes",
			input: []string{"accounts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "EntityDefinitions(LogicalName='account')"):
					_, _ = w.Write(responseContactsSchema)
				case strings.HasSuffix(path, "EntityDefinitions(LogicalName='account')/Attributes"):
					mockutils.WriteBody(w, `{"value":[]}`)
				default:
					_, _ = w.Write([]byte{})
				}
			})),
			expectedErrs: []error{ErrObjectMissingAttributes},
		},
		{
			name:  "Correctly list metadata for account leads and invite contact",
			input: []string{"contacts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// server will be called 2 times
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "EntityDefinitions(LogicalName='contact')"):
					_, _ = w.Write(responseContactsSchema)
				case strings.HasSuffix(path, "EntityDefinitions(LogicalName='contact')/Attributes"):
					_, _ = w.Write(responseContactsAttributes)
				default:
					_, _ = w.Write([]byte{})
				}
			})),
			comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						FieldsMap: map[string]string{
							// nice display names
							"adx_publicprofilecopy":    "Public Profile Copy",
							"adx_identity_newpassword": "New Password Input",
							"department":               "Department",
							"shippingmethodcode":       "Shipping Method",
							"lastname":                 "Last Name",
							// schema name was used for display
							"leadsourcecodename": "LeadSourceCodeName",
							// underscore prefixed fields
							"_accountid_value": "Account",
							"_createdby_value": "Created By",
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
