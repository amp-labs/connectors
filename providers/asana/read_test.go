package asana

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseProjects := testutils.DataFromFile(t, "read-projects.json")
	responseUsers := testutils.DataFromFile(t, "read-users.json")
	responseTags := testutils.DataFromFile(t, "read-tags.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "allocations"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},

		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "attributes", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},

		{
			Name:  "Read list of all projects",
			Input: common.ReadParams{ObjectName: "projects", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProjects),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":           "12345",
							"resource_type": "project",
							"name":          "Stuff to buy",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":           "1245",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all tags",
			Input: common.ReadParams{ObjectName: "tags", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":           "12225",
							"resource_type": "tag",
							"name":          "Stuff to buy",
						},
					},
				},
				Done: true,
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
