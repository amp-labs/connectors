package zohocrm

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	dealResponse := testutils.DataFromFile(t, "deals.json")
	unsupported := testutils.DataFromFile(t, "arsenal.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"deals", "arsenal"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("module", "Deals"),
					Then: mockserver.Response(http.StatusOK, dealResponse),
				}, {
					If:   mockcond.QueryParam("module", "Arsenal"),
					Then: mockserver.Response(http.StatusBadRequest, unsupported),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"deals": {
						DisplayName: "Deals",
						FieldsMap: map[string]string{
							"account_name":           "account_name",
							"amount":                 "amount",
							"campaign_source":        "campaign_source",
							"change_log_time__s":     "change_log_time__s",
							"closing_date":           "closing_date",
							"contact_name":           "contact_name",
							"created_by":             "created_by",
							"created_time":           "created_time",
							"deal_name":              "deal_name",
							"description":            "description",
							"expected_revenue":       "expected_revenue",
							"id":                     "id",
							"last_activity_time":     "last_activity_time",
							"lead_conversion_time":   "lead_conversion_time",
							"lead_source":            "lead_source",
							"locked__s":              "locked__s",
							"modified_by":            "modified_by",
							"modified_time":          "modified_time",
							"next_step":              "next_step",
							"overall_sales_duration": "overall_sales_duration",
							"owner":                  "owner",
							"probability":            "probability",
							"reason_for_loss__s":     "reason_for_loss__s",
							"record_image":           "record_image",
							"record_status__s":       "record_status__s",
							"sales_cycle_duration":   "sales_cycle_duration",
							"stage":                  "stage",
							"type":                   "type",
						},
					},
				},
				Errors: map[string]error{
					"arsenal": mockutils.ExpectedSubsetErrors{
						common.ErrCaller,
						errors.New(string(unsupported)), // nolint:goerr113
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
