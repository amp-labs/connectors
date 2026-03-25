package bentley

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseITwinsFirst := testutils.DataFromFile(t, "read-itwins-first-page.json")
	responseITwinsLast := testutils.DataFromFile(t, "read-itwins-last-page.json")
	responseWebhooks := testutils.DataFromFile(t, "read-webhooks.json")
	responseContextCaptureJobs := testutils.DataFromFile(t, "read-contextcapture-jobs.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "itwins"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unknown object is not supported",
			Input: common.ReadParams{ObjectName: "nonexistent", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK),
			}.Server(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Read itwins first page with next link",
			Input: common.ReadParams{ObjectName: "itwins", Fields: connectors.Fields("id", "displayName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/itwins"),
				Then:  mockserver.Response(http.StatusOK, responseITwinsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "abc-123",
						"displayname": "My Project",
					},
					Raw: map[string]any{
						"id":              "abc-123",
						"class":           "Endeavor",
						"subClass":        "Project",
						"type":            "Bridge",
						"displayName":     "My Project",
						"number":          "PRJ-001",
						"status":          "Active",
						"createdDateTime": "2024-01-15T10:30:00Z",
						"createdBy":       "user@example.com",
					},
					Id: "abc-123",
				}},
				NextPage: "https://api.bentley.com/itwins?$skip=1",
				Done:     false,
			},
		},
		{
			Name:  "Read itwins last page has no next link",
			Input: common.ReadParams{ObjectName: "itwins", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/itwins"),
				Then:  mockserver.Response(http.StatusOK, responseITwinsLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "def-456",
					},
					Raw: map[string]any{
						"id":          "def-456",
						"class":       "Endeavor",
						"subClass":    "Asset",
						"type":        "Road",
						"displayName": "Another Project",
						"status":      "Inactive",
					},
					Id: "def-456",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read webhooks with different response key",
			Input: common.ReadParams{ObjectName: "webhooks", Fields: connectors.Fields("id", "callbackUrl")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/webhooks"),
				Then:  mockserver.Response(http.StatusOK, responseWebhooks),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "wh-789",
						"callbackurl": "https://example.com/webhook",
					},
					Raw: map[string]any{
						"id":          "wh-789",
						"callbackUrl": "https://example.com/webhook",
						"active":      true,
						"eventTypes":  []any{"iModelCreatedEvent"},
						"secret":      "s3cr3t",
					},
					Id: "wh-789",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Incremental read with Since and Until filters",
			Input: common.ReadParams{
				ObjectName: "contextcapture/jobs",
				Fields:     connectors.Fields("id", "name"),
				Since:      time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, time.June, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contextcapture/jobs"),
					mockcond.QueryParam("$filter", "createdDateTime ge 2024-06-01T00:00:00Z and createdDateTime le 2024-06-30T00:00:00Z"), //nolint:lll
				},
				Then: mockserver.Response(http.StatusOK, responseContextCaptureJobs),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "job-001",
						"name": "Scan Bridge",
					},
					Raw: map[string]any{
						"id":              "job-001",
						"name":            "Scan Bridge",
						"iTwinId":         "itwin-abc",
						"email":           "user@example.com",
						"state":           "Completed",
						"createdDateTime": "2024-06-15T10:30:00Z",
					},
					Id: "job-001",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Incremental read with only Since filter",
			Input: common.ReadParams{
				ObjectName: "contextcapture/jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contextcapture/jobs"),
					mockcond.QueryParam("$filter", "createdDateTime ge 2024-06-01T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseContextCaptureJobs),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "job-001",
					},
					Raw: map[string]any{
						"id":    "job-001",
						"name":  "Scan Bridge",
						"state": "Completed",
					},
					Id: "job-001",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Non-incremental object ignores Since filter",
			Input: common.ReadParams{
				ObjectName: "itwins",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/itwins"),
				Then:  mockserver.Response(http.StatusOK, responseITwinsLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "def-456",
					},
					Raw: map[string]any{
						"id":     "def-456",
						"status": "Inactive",
					},
					Id: "def-456",
				}},
				NextPage: "",
				Done:     true,
			},
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
