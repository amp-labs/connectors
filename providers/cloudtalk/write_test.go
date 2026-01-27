package cloudtalk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name: "Create contact successfully",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"name":  "Jane Doe",
					"email": "jane.doe@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/contacts/add.json"),
					mockcond.Body(`{"email":"jane.doe@example.com","name":"Jane Doe"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"responseData": {
						"id": "123",
						"name": "Jane Doe",
						"email": "jane.doe@example.com"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123",
				Data: map[string]any{
					"id":    "123",
					"name":  "Jane Doe",
					"email": "jane.doe@example.com",
				},
			},
		},
		{
			Name: "Update contact successfully",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "123",
				RecordData: map[string]any{
					"name": "Jane Doe Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/contacts/edit/123.json"),
					mockcond.Body(`{"name":"Jane Doe Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"responseData": {
						"id": "123",
						"name": "Jane Doe Updated"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123",
				Data: map[string]any{
					"id":   "123",
					"name": "Jane Doe Updated",
				},
			},
		},
		{
			Name: "Create contact with nested data response",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"name": "Nested User",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/contacts/add.json"),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"responseData": {
						"status": 201,
						"data": {
							"id": 1583318602
						}
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1583318602",
				Data: map[string]any{
					"status": float64(201), // json unmarshals numbers as float64
					"data": map[string]any{
						"id": float64(1583318602),
					},
				},
			},
		},
		{
			Name: "Create tag successfully",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordData: map[string]any{
					"name": "VIP",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/tags/add.json"),
					mockcond.Body(`{"name":"VIP"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"responseData": {
						"id": "555",
						"name": "VIP"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "555",
				Data: map[string]any{
					"id":   "555",
					"name": "VIP",
				},
			},
		},
		{
			Name: "Update tag successfully",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordId:   "555",
				RecordData: map[string]any{
					"name": "VIP Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/tags/edit/555.json"),
					mockcond.Body(`{"name":"VIP Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"responseData": {
						"id": "555",
						"name": "VIP Updated"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "555",
				Data: map[string]any{
					"id":   "555",
					"name": "VIP Updated",
				},
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
