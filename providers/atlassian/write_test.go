package atlassian

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidProjectError := testutils.DataFromFile(t, "create-issue-invalid-project.json")
	responseInvalidTypeError := testutils.DataFromFile(t, "create-issue-invalid-type.json")
	createIssueResponse := testutils.DataFromFile(t, "create-issue.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "issues"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Error missing project during write",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidProjectError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("project:Specify a valid project ID or key"),
			},
		},
		{
			Name:  "Error missing issue type during write",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidTypeError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("issuetype:Specify an issue type"),
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "issues", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of an Issue",
			Input: common.WriteParams{ObjectName: "issues", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createIssueResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "10004",
				Errors:   nil,
				Data:     nil,
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

func TestWriteConfluence(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorInvalidPayload := testutils.DataFromFile(t, "confluence/write/blogposts/bad-request.json")
	responseBlogpost := testutils.DataFromFile(t, "confluence/write/blogposts/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "blogposts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Error bad payload",
			Input: common.WriteParams{ObjectName: "blogposts", RecordId: "1802301", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidPayload),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Non-drafts cannot have empty titles,spaceId: must not be null"), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "blogposts", RecordData: map[string]any{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "blogposts", RecordId: "1802301", RecordData: map[string]any{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12" +
						"/wiki/api/v2/blogposts/1802301"),
					// ID is auto inserted into the payload
					mockcond.Body(`{"id":"1802301"}`),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of an Blogpost",
			Input: common.WriteParams{ObjectName: "blogposts", RecordData: map[string]any{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseBlogpost),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1802301",
				Errors:   nil,
				Data: map[string]any{
					"spaceId":   "131189",
					"createdAt": "2025-07-01T00:27:55.383Z",
					"authorId":  "712020:dcef0c39-e30a-405a-afe5-9d34bf00b73f",
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
				return constructTestConnectorConfluence(tt.Server.URL)
			})
		})
	}
}
