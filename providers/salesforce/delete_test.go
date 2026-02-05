package salesforce

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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "crm/delete/not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Not found returned on removing missing entry",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "10010"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("entity is deleted"),
			},
		},
		{
			Name: "Successful delete",
			Input: common.DeleteParams{
				ObjectName: "contacts",
				RecordId:   "003ak00000luULKAA2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/services/data/v60.0/sobjects/contacts/003ak00000luULKAA2"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestDeletePardot(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorNotFound := testutils.DataFromFile(t, "pardot/delete/err-not-found.json")

	pardotHeader := http.Header{
		"Pardot-Business-Unit-Id": []string{"test-business-unit-id"},
	}

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Successful prospect delete",
			Input: common.DeleteParams{ObjectName: "prosPecTs", RecordId: "55434595"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v5/objects/prospects/55434595"),
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "prosPecTs", RecordId: "55434595"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v5/objects/prospects/55434595"),
				},
				Then: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("The requested record was not found."), // nolint:goerr113
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnectorAccountEngagement(tt.Server.URL)
			})
		})
	}
}
