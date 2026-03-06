package docusign

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

func TestRead(t *testing.T) {
	t.Parallel()

	accountId := "devTest-123"
	accountIdPathPrefix := "/restapi/v2.1/accounts/" + accountId

	responseEnvelopesFirstPage := testutils.DataFromFile(t, "read-envelopes-first-page.json")
	responseEnvelopesSecondPage := testutils.DataFromFile(t, "read-envelopes-second-page.json")

	tests := []testroutines.Read{
		{
			Name: "Read Envelopes first page",
			Input: common.ReadParams{
				ObjectName: "envelopes",
				Fields:     connectors.Fields("envelopeId", "documentsUri", "recipientsUri"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path(accountIdPathPrefix + "/envelopes"),
				Then:  mockserver.Response(http.StatusOK, responseEnvelopesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"envelopeid":    "49a523cb-d283-8f6f-8060-2a04b73f01b7",
						"documentsuri":  "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/documents",
						"recipientsuri": "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/recipients",
					},
					Raw: map[string]any{
						"envelopeId":    "49a523cb-d283-8f6f-8060-2a04b73f01b7",
						"documentsUri":  "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/documents",
						"recipientsUri": "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/recipients",
						"status":        "sent",
						"emailSubject":  "Test Envelope Send 1",
					},
				}},
				NextPage: testroutines.URLTestServer + "/restapi/v2.1/accounts/devTest-123/envelopes?start_position=1&count=1&from_date=2024-03-04T07:22:33-08:00",
				Done:     false,
			},
		},
		{
			Name: "Read Envelopes second page",
			Input: common.ReadParams{
				ObjectName: "envelopes",
				Fields:     connectors.Fields("envelopeId", "documentsUri", "recipientsUri"),
				NextPage:   common.NextPageToken(testroutines.URLTestServer + accountIdPathPrefix + "/envelopes?start_position=1&count=1&from_date=2024-03-04T07:22:33-08:00"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(accountIdPathPrefix + "/envelopes"),
					mockcond.QueryParam("start_position", "1"),
					mockcond.QueryParam("count", "1"),
					mockcond.QueryParam("from_date", "2024-03-04T07:22:33-08:00"),
				},
				Then: mockserver.Response(http.StatusOK, responseEnvelopesSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"envelopeid":    "02da2475-70f8-868f-819f-9fe0dd3f0160",
						"documentsuri":  "/envelopes/02da2475-70f8-868f-819f-9fe0dd3f0160/documents",
						"recipientsuri": "/envelopes/02da2475-70f8-868f-819f-9fe0dd3f0160/recipients",
					},
					Raw: map[string]any{
						"envelopeId":    "02da2475-70f8-868f-819f-9fe0dd3f0160",
						"documentsUri":  "/envelopes/02da2475-70f8-868f-819f-9fe0dd3f0160/documents",
						"recipientsUri": "/envelopes/02da2475-70f8-868f-819f-9fe0dd3f0160/recipients",
						"status":        "sent",
						"emailSubject":  "Test Signing Group 1",
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
				connMetadata := map[string]string{
					"server":     "devTest",
					"account_id": accountId,
				}
				return constructTestConnector(tt.Server.URL, connMetadata)
			})
		})
	}
}
