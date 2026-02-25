package heyreach

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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	campaignResponse := testutils.DataFromFile(t, "campaign.json")
	listResponse := testutils.DataFromFile(t, "list.json")
	liAccountResponse := testutils.DataFromFile(t, "li_account.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all campaign",
			Input: common.ReadParams{ObjectName: "campaign/GetAll", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/public/campaign/GetAll"),
				Then:  mockserver.Response(http.StatusOK, campaignResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": float64(83192),
						},
						Raw: map[string]any{
							"id":           float64(83192),
							"name":         "Test Campaign",
							"creationTime": "2024-12-31T09:17:29.106903Z",
							"status":       "DRAFT",
						},
						Id: "83192",
					},
				},
				NextPage: "100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all linked account",
			Input: common.ReadParams{ObjectName: "li_account/GetAll", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/public/li_account/GetAll"),
				Then:  mockserver.Response(http.StatusOK, liAccountResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": float64(71110),
						},
						Raw: map[string]any{
							"id":           float64(71110),
							"emailAddress": "sample@gmail.com",
							"firstName":    nil,
							"lastName":     nil,
						},
						Id: "71110",
					},
				},
				NextPage: "100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all list",
			Input: common.ReadParams{ObjectName: "list/GetAll", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/public/list/GetAll"),
				Then:  mockserver.Response(http.StatusOK, listResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": float64(188213),
						},
						Raw: map[string]any{
							"id":           float64(188213),
							"name":         "Test 2",
							"listType":     "USER_LIST",
							"creationTime": "2025-03-26T10:24:33.266015Z",
						},
						Id: "188213",
					},
				},
				NextPage: "100",
				Done:     false,
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
