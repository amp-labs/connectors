package msdsales

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
	"github.com/amp-labs/connectors/common/reqrepeater"
	"github.com/go-test/deep"
)

func Test_makeQueryValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    common.ReadParams
		expected string
	}{
		{
			name:     "No params no query string",
			input:    common.ReadParams{},
			expected: "",
		},
		{
			name: "One parameter",
			input: common.ReadParams{
				Fields: []string{"cat"},
			},
			expected: "?$select=cat",
		},
		{
			name: "Many parameters",
			input: common.ReadParams{
				Fields: []string{"cat", "dog", "parrot", "hamster"},
			},
			expected: "?$select=cat,dog,parrot,hamster",
		},
		{
			name: "OData parameters with @ symbol",
			input: common.ReadParams{
				Fields: []string{"cat", "@odata.dog", "parrot"},
			},
			expected: "?$select=cat,@odata.dog,parrot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := makeQueryValues(tt.input)
			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

var contactsFirstPageResponse = `{
		"@odata.context": "https://org5bd08fdd.api.crm.dynamics.com/api/data/v9.2/$metadata#contacts(fullname,emailaddress1,fax,familystatuscode)",
		"@Microsoft.Dynamics.CRM.totalrecordcount": -1,
		"@Microsoft.Dynamics.CRM.totalrecordcountlimitexceeded": false,
		"@Microsoft.Dynamics.CRM.globalmetadataversion": "6012567",
		"value": [
		{
		  "@odata.etag": "W/\"4372108\"",
		  "fullname": "Heriberto Nathan",
		  "emailaddress1": "heriberto@northwindtraders.com",
		  "fax": "614-555-0122",
		  "familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
		  "familystatuscode": 1,
		  "contactid": "cdcfa450-cb0c-ea11-a813-000d3a1b1223"
		},
		{
		  "@odata.etag": "W/\"4372115\"",
		  "fullname": "Dwayne Elijah",
		  "emailaddress1": "dwayne@alpineskihouse.com",
		  "fax": "281-555-0158",
		  "familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
		  "familystatuscode": 1,
		  "contactid": "9fd4a450-cb0c-ea11-a813-000d3a1b1223"
		}
		],
		"@odata.nextLink": "https://org5bd08fdd.api.crm.dynamics.com/api/data/v9.2/contacts?$select=fullname,emailaddress1,fax,familystatuscode&$skiptoken=%3Ccookie%20pagenumber=%222%22%20pagingcookie=%22%253ccookie%2520page%253d%25221%2522%253e%253ccontactid%2520last%253d%2522%257b9FD4A450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520first%253d%2522%257bCDCFA450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520%252f%253e%253c%252fcookie%253e%22%20istracking=%22False%22%20/%3E"
	}`

func Test_Read(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        common.ReadParams
		server       *httptest.Server
		connector    Connector
		expected     *common.ReadResult
		expectedErrs []error
	}{
		{
			name: "Mime response header expected",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.MissingContentType},
		},
		{
			name: "Correct error message is understood from JSON response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeBody(w, `{
					"error": {
						"code": "0x80060888",
						"message":"Resource not found for the segment 'conacs'."
					}
				}`)
			})),
			expectedErrs: []error{common.ErrBadRequest, errors.New("Resource not found for the segment 'conacs'")},
		},
		{
			name: "Incorrect key in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"garbage": {}
				}`)
			})),
			expectedErrs: []error{errors.New("wrong request: wrong key 'value'")},
		},
		{
			name: "Incorrect data type in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"value": {}
				}`)
			})),
			expectedErrs: []error{common.ErrNotArray},
		},
		{
			name: "Next page cursor may be missing in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"value": []
				}`)
			})),
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read with 2 entries",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, contactsFirstPageResponse)
			})),
			expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372108\"",
						"fullname":      "Heriberto Nathan",
						"emailaddress1": "heriberto@northwindtraders.com",
						"fax":           "614-555-0122",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "cdcfa450-cb0c-ea11-a813-000d3a1b1223",
					},
				}, {
					Fields: map[string]any{},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372115\"",
						"fullname":      "Dwayne Elijah",
						"emailaddress1": "dwayne@alpineskihouse.com",
						"fax":           "281-555-0158",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "9fd4a450-cb0c-ea11-a813-000d3a1b1223",
					},
				}},
				NextPage: "https://org5bd08fdd.api.crm.dynamics.com/api/data/v9.2/contacts?$select=fullname,emailaddress1,fax,familystatuscode&$skiptoken=%3Ccookie%20pagenumber=%222%22%20pagingcookie=%22%253ccookie%2520page%253d%25221%2522%253e%253ccontactid%2520last%253d%2522%257b9FD4A450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520first%253d%2522%257bCDCFA450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520%252f%253e%253c%252fcookie%253e%22%20istracking=%22False%22%20/%3E",
				Done:     false,
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read with chosen fields",
			input: common.ReadParams{
				Fields: []string{"fullname", "familystatuscode"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, contactsFirstPageResponse)
			})),
			expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"fullname":         "Heriberto Nathan",
						"familystatuscode": float64(1),
					},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372108\"",
						"fullname":      "Heriberto Nathan",
						"emailaddress1": "heriberto@northwindtraders.com",
						"fax":           "614-555-0122",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "cdcfa450-cb0c-ea11-a813-000d3a1b1223",
					},
				}, {
					Fields: map[string]any{
						"fullname":         "Dwayne Elijah",
						"familystatuscode": float64(1),
					},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372115\"",
						"fullname":      "Dwayne Elijah",
						"emailaddress1": "dwayne@alpineskihouse.com",
						"fax":           "281-555-0158",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "9fd4a450-cb0c-ea11-a813-000d3a1b1223",
					},
				}},
				NextPage: "https://org5bd08fdd.api.crm.dynamics.com/api/data/v9.2/contacts?$select=fullname,emailaddress1,fax,familystatuscode&$skiptoken=%3Ccookie%20pagenumber=%222%22%20pagingcookie=%22%253ccookie%2520page%253d%25221%2522%253e%253ccontactid%2520last%253d%2522%257b9FD4A450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520first%253d%2522%257bCDCFA450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520%252f%253e%253c%252fcookie%253e%22%20istracking=%22False%22%20/%3E",
				Done:     false,
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructin connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.BaseURL = tt.server.URL
			connector.Client.HTTPClient.Base = tt.server.URL
			// failed requests will be not retried
			connector.RetryStrategy = &reqrepeater.NullStrategy{}

			// start of tests
			output, err := connector.Read(ctx, tt.input)
			if len(tt.expectedErrs) == 0 && err != nil {
				t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
			}

			if len(tt.expectedErrs) != 0 && err == nil {
				t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}

func writeBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}
