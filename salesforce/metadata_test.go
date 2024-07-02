package salesforce

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

	responseOrgMeta := mockutils.DataFromFile(t, "organization-metadata.json")

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
			name:  "Mime response header expected for error",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Mime response header expected for successful response",
			input: []string{"butterflies"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{}`)
			})),
			expectedErrs: []error{common.ErrNotJSON},
		},
		{
			name:  "Successfully describe one object with metadata",
			input: []string{"Organization"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToBody(w, r, `{"allOrNone":false,"compositeRequest":[{
					"referenceId":"Organization",
					"method":"GET",
					"url":"/services/data/v59.0/sobjects/Organization/describe"
				}]}`, func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseOrgMeta)
				})
			})),
			expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"organization": {
						DisplayName: "Organization",
						FieldsMap: map[string]string{
							"defaultlocalesidkey":                    "Locale",
							"preferencestransactionsecuritypolicy":   "TransactionSecurityPolicy",
							"lastmodifiedbyid":                       "Last Modified By ID",
							"usesstartdateasfiscalyearname":          "Fiscal Year Name by Start",
							"defaultopportunityaccess":               "Default Opportunity Access",
							"signupcountryisocode":                   "Signup Country",
							"instancename":                           "Instance Name",
							"division":                               "Division",
							"city":                                   "City",
							"languagelocalekey":                      "Language",
							"preferencesonlyllpermuserallowed":       "OnlyLLPermUserAllowed",
							"uiskin":                                 "UI Skin",
							"issandbox":                              "Is Sandbox",
							"createdbyid":                            "Created By ID",
							"defaultcaseaccess":                      "Default Case Access",
							"defaultcampaignaccess":                  "Default Campaign Access",
							"namespaceprefix":                        "Namespace Prefix",
							"lastmodifieddate":                       "Last Modified Date",
							"state":                                  "State/Province",
							"longitude":                              "Longitude",
							"phone":                                  "Phone",
							"receivesadmininfoemails":                "Info Emails Admin",
							"name":                                   "Name",
							"street":                                 "Street",
							"geocodeaccuracy":                        "Geocode Accuracy",
							"timezonesidkey":                         "Time Zone",
							"preferencesconsentmanagementenabled":    "ConsentManagementEnabled",
							"isreadonly":                             "Is Read Only",
							"fax":                                    "Fax",
							"defaultaccountaccess":                   "Default Account Access",
							"monthlypageviewsused":                   "Monthly Page Views Used",
							"primarycontact":                         "Primary Contact",
							"receivesinfoemails":                     "Info Emails",
							"preferencesrequireopportunityproducts":  "RequireOpportunityProducts",
							"organizationtype":                       "Edition",
							"defaultpricebookaccess":                 "Default Price Book Access",
							"systemmodstamp":                         "System Modstamp",
							"country":                                "Country",
							"preferencesautoselectindividualonmerge": "AutoSelectIndividualOnMerge",
							"preferenceslightningloginenabled":       "LightningLoginEnabled",
							"numknowledgeservice":                    "Knowledge Licenses",
							"createddate":                            "Created Date",
							"postalcode":                             "Zip/Postal Code",
							"latitude":                               "Latitude",
							"defaultleadaccess":                      "Default Lead Access",
							"defaultcalendaraccess":                  "Default Calendar Access",
							"webtocasedefaultorigin":                 "Web to Cases Default Origin",
							"monthlypageviewsentitlement":            "Monthly Page Views Allowed",
							"compliancebccemail":                     "Compliance BCC Email",
							"trialexpirationdate":                    "Trial Expiration Date",
							"id":                                     "Organization ID",
							"address":                                "Address",
							"fiscalyearstartmonth":                   "Fiscal Year Starts In",
							"defaultcontactaccess":                   "Default Contact Access",
						},
					},
				},
				Errors: map[string]error{},
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
