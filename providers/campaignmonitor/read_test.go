package campaignmonitor

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	clientsResponse := testutils.DataFromFile(t, "clients.json")
	adminsResponse := testutils.DataFromFile(t, "admins.json")
	campaignsResponse := testutils.DataFromFile(t, "campaigns.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all clients",
			Input: common.ReadParams{ObjectName: "clients", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, clientsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"ClientID": "4a397ccaaa55eb4e6aa1221e1e2d7122",
							"Name":     "Client One",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all admins",
			Input: common.ReadParams{ObjectName: "admins", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, adminsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"EmailAddress": "sally@sparrow.com",
							"Name":         "Sally Sparrow",
							"Status":       "Waiting to Accept the Invitation",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read list of all campaigns",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields(""),
				NextPage:   "",
				Since:      time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, campaignsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"Name":              "First Campaign",
							"FromName":          "sample",
							"FromEmail":         "sample@gmail.com",
							"ReplyTo":           "sample@gmail.com",
							"SentDate":          "2024-08-19 09:35:00",
							"TotalRecipients":   float64(2),
							"CampaignID":        "90be62122fdb35bf09e2a0030aa0b92c",
							"Subject":           "qwtrwre",
							"Tags":              []any{},
							"WebVersionURL":     "http://createsend.com/t/y-3E07774FA5C0962D2540EF23F30FEDED",
							"WebVersionTextURL": "http://createsend.com/t/y-3E07774FA5C0962D2540EF23F30FEDED/t",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/api/v3.3/clients/744cdce058fc61d9ef5e2492f8d8fbaf/campaigns.json?" +
					"pageSize=1000&sentFromDate=2024-05-01&sentToDate=2024-10-01&page=2",
				Done: false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
