package outreach

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "prospects.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	callsResponse := testutils.DataFromFile(t, "calls.json")
	mailingsResponse := testutils.DataFromFile(t, "mailings.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "arsenal", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				errors.New(string(unsupportedResponse)), //nolint:err113
			},
		},
		{
			Name:  "Forbidden access object",
			Input: common.ReadParams{ObjectName: "forbidden", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, callsResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				errors.New(string(callsResponse)), //nolint:err113
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "mailboxes", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Mailings",
			Input: common.ReadParams{
				ObjectName: "mailings",
				Fields:     connectors.Fields("bodyHtml", "errorReason", "id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, mailingsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"bodyhtml":    "\u003chtml\u003e\u003cbody\u003e\u003cp\u003eHere Goes your HTML email\u003c/p\u003e\u003c/body\u003e\u003e\u003c/html\u003e", //nolint:lll
						"errorreason": nil,
						"id":          1,
					},
					Raw: map[string]any{
						"bodyHtml":               "\u003chtml\u003e\u003cbody\u003e\u003cp\u003eHere Goes your HTML email\u003c/p\u003e\u003c/body\u003e\u003e\u003c/html\u003e", //nolint:lll
						"bodyText":               "Here Goes your HTML email\u003e",
						"clickCount":             float64(0),
						"errorreason":            nil,
						"createdAt":              "2024-07-26T06:27:17.000Z",
						"followUpTaskType":       "string",
						"mailboxAddress":         "willy@withampersand.com",
						"mailingType":            "sequence",
						"id":                     1,
						"openCount":              float64(0),
						"overrideSafetySettings": false,
						"references":             []any{},
						"retryCount":             float64(0),
						"scheduledAt":            "2019-08-24T14:15:22.000Z",
						"state":                  "drafted",
						"stateChangedAt":         "2024-07-26T06:27:17.000Z",
						"subject":                "My Email Subject",
						"trackLinks":             true,
						"trackOpens":             true,
						"updatedAt":              "2024-08-02T22:28:21.000Z",
					},
				}},
				NextPage: "",
				Done:     true,
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
