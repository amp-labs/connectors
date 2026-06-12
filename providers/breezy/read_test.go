package breezy

import (
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

const testCompanyID = "abc123def456"

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseCompanies := testutils.DataFromFile(t, "read/companies.json")
	responsePositions := testutils.DataFromFile(t, "read/positions.json")
	responseWebhookEndpoints := testutils.DataFromFile(t, "read/webhook-endpoints.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectCompanies},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Read companies",
			Input: common.ReadParams{ObjectName: objectCompanies, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/companies"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "abc123def456",
						"name": "Acme Corp",
					},
					Raw: map[string]any{
						"_id":          "abc123def456",
						"name":         "Acme Corp",
						"friendly_id":  "acme",
						"initial":      "A",
						"member_count": float64(5),
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read positions requires company_id metadata",
			Input: common.ReadParams{ObjectName: objectPositions, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/positions"),
				},
				Then: mockserver.Response(http.StatusOK, responsePositions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "pos001",
						"name": "Software Engineer",
					},
					Raw: map[string]any{
						"_id":        "pos001",
						"name":       "Software Engineer",
						"type":       "fullTime",
						"state":      "published",
						"department": "Engineering",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read webhook endpoints",
			Input: common.ReadParams{ObjectName: objectWebhookEndpoints, Fields: connectors.Fields("id", "url")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/webhook_endpoints"),
				},
				Then: mockserver.Response(http.StatusOK, responseWebhookEndpoints),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":  "wh001",
						"url": "https://example.com/webhook",
					},
					Raw: map[string]any{
						"id":          "wh001",
						"url":         "https://example.com/webhook",
						"description": "Production webhook",
						"status":      "active",
						"enabled":     true,
					},
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

func TestReadPositionsMissingCompanyID(t *testing.T) {
	t.Parallel()

	tt := testroutines.Read{
		Name:         "Read positions without company_id returns error",
		Input:        common.ReadParams{ObjectName: objectPositions, Fields: connectors.Fields("_id")},
		Server:       mockserver.Dummy(),
		ExpectedErrs: []error{ErrMissingCompanyID},
	}

	tt.Run(t, func() (connectors.ReadConnector, error) {
		return constructTestConnectorWithoutCompanyID(tt.Server.URL)
	})
}

func constructTestConnectorWithoutCompanyID(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
