package freshchat

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createdUser := []byte(`{"id": "u-1", "first_name": "Milton", "email": "milton@example.com"}`)

	updatedAgent := []byte(`{"id": "a-1", "biography": "Support lead", "email": "agent@example.com"}`)
	agentStatus := []byte(`{
		"availability_status": "UNAVAILABLE",
		"agent_status": {"id": "s-1", "name": "On a break"}
	}`)
	createdConversation := []byte(`{"conversation_id": "c-1", "channel_id": "ch-1", "status": "new"}`)
	outboundMessage := []byte(`{"request_id": "r-1", "request_process_time": "30"}`)

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Create posts to the collection and returns the new record",
			Input: common.WriteParams{ObjectName: "users", RecordData: map[string]any{"first_name": "Milton"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/users"),
				},
				Then: mockserver.Response(http.StatusCreated, createdUser),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "u-1",
				Data: map[string]any{
					"id":         "u-1",
					"first_name": "Milton",
					"email":      "milton@example.com",
				},
			},
		},
		{
			// Freshchat answers an update with no body, so the identifier
			// comes from the record that was addressed.
			Name: "Update puts to the record and returns its identifier",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "u-1",
				RecordData: map[string]any{"last_name": "Doe"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/users/u-1"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "u-1",
			},
		},
		{
			Name: "Agent update puts to the record",
			Input: common.WriteParams{
				ObjectName: "agents",
				RecordId:   "a-1",
				RecordData: map[string]any{"biography": "Support lead"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/agents/a-1"),
				},
				Then: mockserver.Response(http.StatusOK, updatedAgent),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a-1",
				Data:     map[string]any{"id": "a-1", "biography": "Support lead"},
			},
		},
		{
			// Freshchat changes the status of an agent with PATCH. The answer does not
			// name the agent, so the identifier comes from the record that was addressed.
			Name: "Agent status update patches the record",
			Input: common.WriteParams{
				ObjectName: "agents",
				RecordId:   "a-1",
				RecordData: map[string]any{
					"agent_status": map[string]any{"id": "s-1", "name": "On a break"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/agents/a-1"),
					mockcond.Body(`{"agent_status": {"id": "s-1", "name": "On a break"}}`),
				},
				Then: mockserver.Response(http.StatusOK, agentStatus),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a-1",
				Data:     map[string]any{"availability_status": "UNAVAILABLE"},
			},
		},
		{
			// A status sent along with other properties is a regular update.
			Name: "Agent status with other properties puts to the record",
			Input: common.WriteParams{
				ObjectName: "agents",
				RecordId:   "a-1",
				RecordData: map[string]any{
					"biography":    "Support lead",
					"agent_status": map[string]any{"id": "s-1", "name": "On a break"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/agents/a-1"),
				},
				Then: mockserver.Response(http.StatusOK, updatedAgent),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a-1",
				Data:     map[string]any{"id": "a-1"},
			},
		},
		{
			// Freshchat names the identifier of a conversation `conversation_id`.
			Name: "Conversation create returns its identifier",
			Input: common.WriteParams{
				ObjectName: "conversations",
				RecordData: map[string]any{"channel_id": "ch-1"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/conversations"),
				},
				Then: mockserver.Response(http.StatusAccepted, createdConversation),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "c-1",
				Data:     map[string]any{"conversation_id": "c-1", "status": "new"},
			},
		},
		{
			Name: "Outbound WhatsApp message posts to its collection",
			Input: common.WriteParams{
				ObjectName: "outbound-messages/whatsapp",
				RecordData: map[string]any{"provider": "whatsapp"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/outbound-messages/whatsapp"),
				},
				Then: mockserver.Response(http.StatusAccepted, outboundMessage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "r-1",
				Data:     map[string]any{"request_id": "r-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
