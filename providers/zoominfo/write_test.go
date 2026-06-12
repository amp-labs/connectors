package zoominfo

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

func TestWrite(t *testing.T) { // nolint:funlen
	t.Parallel()

	audienceResponse := testutils.DataFromFile(t, "write-audience.json")
	buyerPersonaResponse := testutils.DataFromFile(t, "write-buyer-persona.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Input:        common.WriteParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unsupported (read-only) object is rejected",
			Input:        common.WriteParams{ObjectName: objIndustries, RecordData: map[string]any{}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name: "Studio create: POST collection (audiences)",
			Input: common.WriteParams{
				ObjectName: objAudiences,
				RecordData: map[string]any{"name": "Q1 Prospects", "type": "CONTACT", "origin": "CUSTOM"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/studio/v1/audiences"),
					mockcond.MethodPOST(),
					mockcond.Body(
						`{"data":{"type":"Audience","attributes":` +
							`{"name":"Q1 Prospects","origin":"CUSTOM","type":"CONTACT"}}}`,
					),
				},
				Then: mockserver.Response(http.StatusOK, audienceResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "550e8400-e29b-41d4-a716-446655440000",
				Data: map[string]any{
					"id":   "550e8400-e29b-41d4-a716-446655440000",
					"type": typeAudience,
					"attributes": map[string]any{
						"name": "Q1 Prospects", "type": "CONTACT", "recordCount": float64(0), "origin": "CUSTOM",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Studio update: PATCH collection/{id} (audiences)",
			Input: common.WriteParams{
				ObjectName: objAudiences,
				RecordId:   "550e8400-e29b-41d4-a716-446655440000",
				RecordData: map[string]any{"name": "Renamed"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/studio/v1/audiences/550e8400-e29b-41d4-a716-446655440000"),
					mockcond.MethodPATCH(),
					mockcond.Body(`{"data":{"type":"Audience","attributes":{"name":"Renamed"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, audienceResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "550e8400-e29b-41d4-a716-446655440000",
				Data: map[string]any{
					"id":   "550e8400-e29b-41d4-a716-446655440000",
					"type": typeAudience,
					"attributes": map[string]any{
						"name": "Q1 Prospects", "type": "CONTACT", "recordCount": float64(0), "origin": "CUSTOM",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Copilot upsert create: POST collection without id (customer-buyer-personas)",
			Input: common.WriteParams{
				ObjectName: objCustomerBuyerPersonas,
				RecordData: map[string]any{"name": "VP of Engineering"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/copilot/v1/customer-buyer-personas"),
					mockcond.MethodPOST(),
					mockcond.Body(`{"data":{"type":"CustomerBuyerPersona","attributes":{"name":"VP of Engineering"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, buyerPersonaResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "bp_12345",
				Data: map[string]any{
					"id":   "bp_12345",
					"type": typeCustomerBuyerPersona,
					"attributes": map[string]any{
						"name": "VP of Engineering", "description": "Technical decision maker",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Copilot upsert update: POST collection with id in body (customer-buyer-personas)",
			Input: common.WriteParams{
				ObjectName: objCustomerBuyerPersonas,
				RecordId:   "bp_12345",
				RecordData: map[string]any{"description": "Updated"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/copilot/v1/customer-buyer-personas"),
					mockcond.MethodPOST(),
					mockcond.Body(
						`{"data":{"type":"CustomerBuyerPersona","id":"bp_12345",` +
							`"attributes":{"description":"Updated"}}}`,
					),
				},
				Then: mockserver.Response(http.StatusOK, buyerPersonaResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "bp_12345",
				Data: map[string]any{
					"id":   "bp_12345",
					"type": typeCustomerBuyerPersona,
					"attributes": map[string]any{
						"name": "VP of Engineering", "description": "Technical decision maker",
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
