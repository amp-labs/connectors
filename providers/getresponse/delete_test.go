package getresponse

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	// Sample error response for 404 Not Found
	errorResponseNotFound := `{
		"httpStatus": 404,
		"message": "Resource not found",
		"context": []
	}`

	// Sample error response for 400 Bad Request
	errorResponseBadRequest := `{
		"httpStatus": 400,
		"message": "Bad Request",
		"context": [
			{
				"field": "recordId",
				"message": "Invalid record ID format"
			}
		]
	}`

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete record ID must be included",
			Input:        common.DeleteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Delete contact successfully (204 No Content)",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "pV3r"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/contacts/pV3r"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete contact successfully (200 OK)",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "pV4s"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/contacts/pV4s"),
				},
				Then: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete campaign successfully",
			Input: common.DeleteParams{ObjectName: "campaigns", RecordId: "f4PSi"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/campaigns/f4PSi"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Error 404 Not Found",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "nonexistent"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, []byte(errorResponseNotFound)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				errors.New("retryable error"),
			},
		},
		{
			Name:  "Error 400 Bad Request",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "invalid-id"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, []byte(errorResponseBadRequest)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New("caller error"),
			},
		},
		{
			Name:  "Error 401 Unauthorized",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "pV3r"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, []byte(`{"message": "Unauthorized"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				errors.New(`HTTP status 401: access token invalid: {"message": "Unauthorized"}`),
			},
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
