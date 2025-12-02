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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createClientsResponse := testutils.DataFromFile(t, "create-clients.json")
	createAppointmentsResponse := testutils.DataFromFile(t, "create-appointments.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name: "Successfully creation of a client",
			Input: common.WriteParams{ObjectName: "clients", RecordData: map[string]any{
				"firstName": "Bob",
				"lastName":  "Burger",
				"phone":     "555-555-5555",
				"notes":     "note test",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createClientsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"firstName": "Bob",
					"lastName":  "Burger",
					"phone":     "555-555-5555",
					"notes":     "note test",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully creation of an appointment",
			Input: common.WriteParams{ObjectName: "appointments", RecordData: map[string]any{
				"firstName": "Bob",
				"lastName":  "McTest",
				"email":     "bob.mctest@example.com",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createAppointmentsResponse),
			}.Server(),

			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "31991639",
				Data: map[string]any{
					"id":                float64(31991639),
					"firstName":         "Bob",
					"lastName":          "McTest",
					"phone":             "",
					"email":             "bob.mctest@example.com",
					"date":              "February 3, 2016",
					"time":              "2:00pm",
					"endTime":           "3:00pm",
					"dateCreated":       "February 2, 2016",
					"datetime":          "2016-02-03T14:00:00-0800",
					"price":             "0.00",
					"paid":              "no",
					"amountPaid":        "0.00",
					"type":              "Regular Visit",
					"appointmentTypeID": float64(1),
					"classID":           nil,
					"category":          "",
					"duration":          "60",
					"calendar":          "My Calendar",
					"calendarID":        float64(1),
					"location":          "",
					"certificate":       "ABC123",
					"confirmationPage":  "https://www.acuityscheduling.com/schedule.php",
					"formsText":         "...",
					"notes":             "",
					"timezone":          "America/Los_Angeles",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
