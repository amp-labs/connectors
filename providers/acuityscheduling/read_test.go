package acuityscheduling

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	clientsResponse := testutils.DataFromFile(t, "clients-read.json")
	blocksResponse := testutils.DataFromFile(t, "blocks-read.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "clients"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successfully read clients",
			Input: common.ReadParams{ObjectName: "clients", Fields: connectors.Fields("firstName", "lastName", "email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v1/clients"),
				Then:  mockserver.Response(http.StatusOK, clientsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"firstname": "Dipu",
							"lastname":  "Chaurasiya",
							"email":     "dipu.chaurasiya91@gmail.com",
						},
						Raw: map[string]any{
							"firstName": "Dipu",
							"lastName":  "Chaurasiya",
							"email":     "dipu.chaurasiya91@gmail.com",
							"phone":     "+9779807202692",
							"notes":     "",
						},
					},
					{
						Fields: map[string]any{
							"firstname": "Jane",
							"lastname":  "McTest",
							"email":     "jane.mctest@example.com",
						},
						Raw: map[string]any{
							"firstName": "Jane",
							"lastName":  "McTest",
							"email":     "jane.mctest@example.com",
							"phone":     "(123) 555-0101",
							"notes":     "",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read blocks",
			Input: common.ReadParams{ObjectName: "blocks", Fields: connectors.Fields("id", "description", "calendarID")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v1/blocks"),
				Then:  mockserver.Response(http.StatusOK, blocksResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          float64(4),
							"description": "Every Wednesday 11:00am to 12:00pm starting  8 July 2015 ending 29 July 2015",
							"calendarid":  float64(1),
						},
						Raw: map[string]any{
							"description": "Every Wednesday 11:00am to 12:00pm starting  8 July 2015 ending 29 July 2015",
							"until":       "2015-07-29T12:00:00-0700",
							"recurring":   "weekly",
							"notes":       "Recurring blocked time.",
							"end":         "2015-07-08T12:00:00-0700",
							"start":       "2015-07-08T11:00:00-0700",
							"calendarID":  float64(1),
							"id":          float64(4),
						},
					},
					{
						Fields: map[string]any{
							"id":          float64(1),
							"description": "Wednesday  1 July 2015 11:00am - 12:00pm",
							"calendarid":  float64(1),
						},
						Raw: map[string]any{
							"description": "Wednesday  1 July 2015 11:00am - 12:00pm",
							"until":       nil,
							"recurring":   nil,
							"notes":       "Blocked time.",
							"end":         "2015-07-01T12:00:00-0700",
							"start":       "2015-07-01T11:00:00-0700",
							"calendarID":  float64(1),
							"id":          float64(1),
						},
					},
					{
						Fields: map[string]any{
							"id":          float64(3),
							"description": "Tuesday 30 June 11:00am -  1 Jul 2015 12:00pm",
							"calendarid":  float64(2),
						},
						Raw: map[string]any{
							"description": "Tuesday 30 June 11:00am -  1 Jul 2015 12:00pm",
							"until":       nil,
							"recurring":   nil,
							"notes":       "Blocked time.",
							"end":         "2015-07-01T12:00:00-0700",
							"start":       "2015-06-30T11:00:00-0700",
							"calendarID":  float64(2),
							"id":          float64(3),
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
