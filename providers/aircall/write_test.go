package aircall

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Write{
		// --- Contacts (POST/POST) ---
		{
			Name: "Create contact successfully",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"first_name": "John",
					"last_name":  "Doe",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/contacts"),
					mockcond.Body(`{"first_name":"John","last_name":"Doe"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"contact": {
						"id": 12345,
						"first_name": "John",
						"last_name": "Doe"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"id":         float64(12345),
					"first_name": "John",
					"last_name":  "Doe",
				},
			},
		},
		{
			Name: "Update contact successfully",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "12345",
				RecordData: map[string]any{
					"first_name": "Jane",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost), // Aircall uses POST for contact updates
					mockcond.Path("/v1/contacts/12345"),
					mockcond.Body(`{"first_name":"Jane"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"contact": {
						"id": 12345,
						"first_name": "Jane",
						"last_name": "Doe"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"id":         float64(12345),
					"first_name": "Jane",
					"last_name":  "Doe",
				},
			},
		},

		// --- Users (POST/PUT) ---
		{
			Name: "Create user successfully",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordData: map[string]any{
					"email":      "test@example.com",
					"first_name": "Test",
					"last_name":  "User",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/users"),
					mockcond.Body(`{"email":"test@example.com","first_name":"Test","last_name":"User"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"user": {
						"id": 555,
						"email": "test@example.com",
						"name": "Test User"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "555",
				Data: map[string]any{
					"id":    float64(555),
					"email": "test@example.com",
					"name":  "Test User",
				},
			},
		},
		{
			Name: "Update user successfully",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "555",
				RecordData: map[string]any{
					"first_name": "Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut), // Users use PUT for updates
					mockcond.Path("/v1/users/555"),
					mockcond.Body(`{"first_name":"Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"user": {
						"id": 555,
						"name": "Updated User"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "555",
				Data: map[string]any{
					"id":   float64(555),
					"name": "Updated User",
				},
			},
		},

		// --- Tags (POST/PUT) ---
		{
			Name: "Create tag successfully",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordData: map[string]any{
					"name":  "VIP",
					"color": "#FF0000",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/tags"),
					mockcond.Body(`{"color":"#FF0000","name":"VIP"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"tag": {
						"id": 10,
						"name": "VIP",
						"color": "#FF0000"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "10",
				Data: map[string]any{
					"id":    float64(10),
					"name":  "VIP",
					"color": "#FF0000",
				},
			},
		},
		{
			Name: "Update tag successfully",
			Input: common.WriteParams{
				ObjectName: "tags",
				RecordId:   "10",
				RecordData: map[string]any{
					"color": "#00FF00",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut), // Tags use PUT for updates
					mockcond.Path("/v1/tags/10"),
					mockcond.Body(`{"color":"#00FF00"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"tag": {
						"id": 10,
						"name": "VIP",
						"color": "#00FF00"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "10",
				Data: map[string]any{
					"id":    float64(10),
					"name":  "VIP",
					"color": "#00FF00",
				},
			},
		},

		// --- Teams (POST/PUT) ---
		{
			Name: "Create team successfully",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordData: map[string]any{
					"name": "Support",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/teams"),
					mockcond.Body(`{"name":"Support"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"team": {
						"id": 88,
						"name": "Support"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "88",
				Data: map[string]any{
					"id":   float64(88),
					"name": "Support",
				},
			},
		},
		{
			Name: "Update team successfully",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordId:   "88",
				RecordData: map[string]any{"name": "New Name"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/v1/teams/88"),
					mockcond.Body(`{"name":"New Name"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"team": {
						"id": 88,
						"name": "New Name"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "88",
				Data: map[string]any{
					"id":   float64(88),
					"name": "New Name",
				},
			},
		},

		// --- Numbers (POST/PUT) ---
		{
			Name: "Create number successfully",
			Input: common.WriteParams{
				ObjectName: "numbers",
				RecordData: map[string]any{
					"digits": "+1234567890",
					"name":   "Main Line",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/numbers"),
					mockcond.Body(`{"digits":"+1234567890","name":"Main Line"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"number": {
						"id": 123,
						"digits": "+1234567890",
						"name": "Main Line"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123",
				Data: map[string]any{
					"id":     float64(123),
					"digits": "+1234567890",
					"name":   "Main Line",
				},
			},
		},
		{
			Name: "Update number successfully",
			Input: common.WriteParams{
				ObjectName: "numbers",
				RecordId:   "123",
				RecordData: map[string]any{
					"name": "Updated Line",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut), // Numbers use PUT for updates
					mockcond.Path("/v1/numbers/123"),
					mockcond.Body(`{"name":"Updated Line"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"number": {
						"id": 123,
						"name": "Updated Line"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123",
				Data: map[string]any{
					"id":   float64(123),
					"name": "Updated Line",
				},
			},
		},
		{
			Name: "Create call successfully",
			Input: common.WriteParams{
				ObjectName: "calls",
				RecordData: map[string]any{
					"number": "+1234567890",
					"to":     "+0987654321",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/v1/calls"),
					mockcond.Body(`{"number":"+1234567890","to":"+0987654321"}`),
				},
				Then: mockserver.Response(http.StatusCreated, []byte(`{
					"call": {
						"id": 999,
						"status": "dialing"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "999",
				Data: map[string]any{
					"id":     float64(999),
					"status": "dialing",
				},
			},
		},

		// --- Error Handling ---
		{
			Name: "Unauthorized error",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{"first_name": "John"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, []byte(`{"error": "Unauthorized"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("Unauthorized"),
			},
		},
		{
			Name: "Bad request error",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{}, // Missing required fields
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, []byte(`{"error": "Missing required fields"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				testutils.StringError("Missing required fields"),
			},
		},
		{
			Name: "Internal server error",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{"first_name": "John"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, []byte(`{"error": "Internal error"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrServer,
				testutils.StringError("Internal error"),
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
