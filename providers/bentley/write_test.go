package bentley

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCreateITwin := testutils.DataFromFile(t, "write-create-itwin.json")
	responseUpdateITwin := testutils.DataFromFile(t, "write-update-itwin.json")
	responseCreateWebhook := testutils.DataFromFile(t, "write-create-webhook.json")
	responseUpdateManufacturer := testutils.DataFromFile(t, "write-update-manufacturer.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "itwins"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Unsupported object returns error",
			Input: common.WriteParams{ObjectName: "curated-content/cesium", RecordData: map[string]any{"key": "val"}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK),
			}.Server(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Successfully create an iTwin",
			Input: common.WriteParams{
				ObjectName: "itwins",
				RecordData: map[string]any{
					"class":       "Endeavor",
					"subClass":    "Project",
					"type":        "Bridge",
					"displayName": "New Bridge Project",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/itwins"),
				},
				Then: mockserver.Response(http.StatusCreated, responseCreateITwin),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Data: map[string]any{
					"iTwin": map[string]any{
						"id":              "itwin-new-001",
						"class":           "Endeavor",
						"subClass":        "Project",
						"type":            "Bridge",
						"displayName":     "New Bridge Project",
						"number":          "PRJ-100",
						"status":          "Active",
						"createdDateTime": "2024-06-15T10:30:00Z",
						"createdBy":       "user@example.com",
					},
				},
			},
		},
		{
			Name: "Successfully update an iTwin via PATCH",
			Input: common.WriteParams{
				ObjectName: "itwins",
				RecordId:   "itwin-existing-001",
				RecordData: map[string]any{
					"displayName": "Updated Road Project",
					"type":        "Road",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/itwins/itwin-existing-001"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateITwin),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Data: map[string]any{
					"iTwin": map[string]any{
						"id":              "itwin-existing-001",
						"class":           "Endeavor",
						"subClass":        "Project",
						"type":            "Road",
						"displayName":     "Updated Road Project",
						"number":          "PRJ-200",
						"status":          "Active",
						"createdDateTime": "2024-01-10T08:00:00Z",
						"createdBy":       "user@example.com",
					},
				},
			},
		},
		{
			Name: "Successfully create a webhook with flat response",
			Input: common.WriteParams{
				ObjectName: "webhooks",
				RecordData: map[string]any{
					"callbackUrl": "https://example.com/webhook",
					"scope":       "iTwin",
					"scopeId":     "scope-123",
					"eventTypes":  []string{"iModelCreatedEvent"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/webhooks"),
				},
				Then: mockserver.Response(http.StatusCreated, responseCreateWebhook),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "wh-new-001",
				Data: map[string]any{
					"id":          "wh-new-001",
					"callbackUrl": "https://example.com/webhook",
					"scope":       "iTwin",
					"scopeId":     "scope-123",
					"active":      true,
					"eventTypes":  []any{"iModelCreatedEvent"},
					"secret":      "s3cr3t",
				},
			},
		},
		{
			Name: "Library objects use PUT for updates instead of PATCH",
			Input: common.WriteParams{
				ObjectName: "library/manufacturers",
				RecordId:   "mfr-001",
				RecordData: map[string]any{
					"displayName": "Updated Manufacturer",
					"description": "Updated description",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/library/manufacturers/mfr-001"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateManufacturer),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"manufacturer": map[string]any{
						"id":          "mfr-001",
						"displayName": "Updated Manufacturer",
						"description": "Updated description",
					},
				},
			},
		},
		{
			Name: "Write with no response body returns success",
			Input: common.WriteParams{
				ObjectName: "itwins/favorites",
				RecordId:   "fav-123",
				RecordData: map[string]any{"id": "fav-123"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/itwins/favorites/fav-123"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
