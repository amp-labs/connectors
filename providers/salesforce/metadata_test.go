package salesforce

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseOrgMeta := testutils.DataFromFile(t, "organization-metadata.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Mime response header expected for successful response",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Always: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			ExpectedErrs: []error{common.ErrNotJSON},
		},
		{
			Name:  "Successfully describe one object with metadata",
			Input: []string{"Organization"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"allOrNone":false,"compositeRequest":[{
					"referenceId":"Organization",
					"method":"GET",
					"url":"/services/data/v59.0/sobjects/Organization/describe"
				}]}`),
				Then: mockserver.Response(http.StatusOK, responseOrgMeta),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
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
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
