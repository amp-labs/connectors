package klaviyo

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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorUnauthenticated := testutils.DataFromFile(t, "invalid-auth.json")
	errorTagBadRequest := testutils.DataFromFile(t, "write-tag-bad-request.json")
	errorTagDuplicate := testutils.DataFromFile(t, "write-tag-conflict.json")
	responseCreateTag := testutils.DataFromFile(t, "write-tag.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "tags"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Unauthenticated error",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordId:   "9891d452-56fe-4397-b431-a92e79cdc980",
				RecordData: "dummy",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/vnd.api+json"),
				Always: mockserver.Response(http.StatusUnauthorized, errorUnauthenticated),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				errors.New( // nolint:goerr113
					"Incorrect authentication credentials.",
				),
			},
		},
		{
			Name: "Error on bad tag payload",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordId:   "9891d452-56fe-4397-b431-a92e79cdc980",
				RecordData: "dummy",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/vnd.api+json"),
				Always: mockserver.Response(http.StatusBadRequest, errorTagBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Invalid input: One of `attributes`, `relationships` or `id` must be included in the request payload.", // nolint:lll
				),
			},
		},
		{
			Name: "Error on duplicate tag creation",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordId:   "9891d452-56fe-4397-b431-a92e79cdc980",
				RecordData: "dummy",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/vnd.api+json"),
				Always: mockserver.Response(http.StatusConflict, errorTagDuplicate),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Tag Service Error: Tag with name 'popular' already exists"), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Write must act as an Update",
			Input: common.WriteParams{
				ObjectName: "campaigns",
				RecordId:   "01JCPFHB29QZ1NDPR3GCGQS5G2",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a tag",
			Input: common.WriteParams{ObjectName: "tags", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateTag),
			}.Server(),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9891d452-56fe-4397-b431-a92e79cdc980",
				Errors:   nil,
				Data: map[string]any{
					"id": "9891d452-56fe-4397-b431-a92e79cdc980",
					"attributes": map[string]any{
						"name": "popular",
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

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
