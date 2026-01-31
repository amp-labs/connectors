package salesfinity

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

func TestWrite(t *testing.T) {
	t.Parallel()

	responseContactListsWrite := testutils.DataFromFile(t, "contact-lists-write-response.json")

	tests := []testroutines.Write{
		{
			Name: "Create contact list successfully",
			Input: common.WriteParams{
				ObjectName: "contact-lists",
				RecordData: map[string]any{
					"name":    "Test Contact List",
					"user_id": "695ee359f1bbcd2c51d4ae1a",
					"contacts": []any{
						map[string]any{
							"first_name": "John",
							"last_name":  "Doe",
							"email":      "john.doe@example.com",
							"company":    "Example Corp",
							"title":      "Software Engineer",
							"phone_numbers": []any{
								map[string]any{
									"type":         "mobile",
									"number":       "5551234567",
									"country_code": "+1",
								},
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/contact-lists"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactListsWrite),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6972b6679feab382af08f409",
				Data: map[string]any{
					"_id":  "6972b6679feab382af08f409",
					"name": "Test Contact List",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create contact list with minimal payload",
			Input: common.WriteParams{
				ObjectName: "contact-lists",
				RecordData: map[string]any{
					"name":     "Minimal List",
					"user_id":  "user123",
					"contacts": []any{},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/contact-lists"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"_id": "minimal_list_id_123",
					"name": "Minimal List",
					"user": "user123",
					"contacts": [],
					"createdAt": "2026-01-22T23:44:39.646Z",
					"updatedAt": "2026-01-22T23:44:39.646Z"
				}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "minimal_list_id_123",
				Data: map[string]any{
					"_id":  "minimal_list_id_123",
					"name": "Minimal List",
				},
			},
			ExpectedErrs: nil,
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
