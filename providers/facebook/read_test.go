package facebook

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	usersResponse := testutils.DataFromFile(t, "users.json")
	adimagesResponse := testutils.DataFromFile(t, "adimages.json")
	systemUsersResponse := testutils.DataFromFile(t, "system_users.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"name": "Ampersand Ampersand",
							"tasks": []any{
								"DRAFT",
								"ANALYZE",
								"ADVERTISE",
								"MANAGE",
							},
							"id": "122142688934782286",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all adimages",
			Input: common.ReadParams{ObjectName: "adimages", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, adimagesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"hash": "29435d869636b1ae9e38480f4aca7f74",
							"id":   "1214321106978726:29435d869636b1ae9e38480f4aca7f74",
						},
					},
				},
				NextPage: "https://graph.facebook.com/v19.0/act_1214321106978726/adimages?limit=25" +
					"&after=QVFIUkdfVFFscVJoNC04b3JYdnFWUkdpTUUzeHhYc2ZABRi1IX05PRkxhWkVqNVJfUG5tb3p1aEVMS2JEQktfR0FhMHpvaFZAubEpNTlFKUkZAsZAWdqeHZAvQUxR",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all system users",
			Input: common.ReadParams{ObjectName: "system_users", Fields: connectors.Fields(""), NextPage: ""},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, systemUsersResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":   "122096363306901679",
							"name": "My System User",
							"role": "ADMIN",
						},
					},
				},
				NextPage: "https://graph.facebook.com/v19.0/1190021932394709/system_users?limit=25" +
					"&after=QVFIUkVWY19yMHBzbmYzNXBSTlBuc1hxU0ZARbjh6dy1LMGFlNHJTM1RPVTFWTExCQ1V2aWRxZAW9UX3kxS29NZAndmeFpEX0JBMUhsZAlQ3dDZAxSFU3VnlCNGVn",
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
