package docusign

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
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	t.Parallel()

	accountId := "devTest-123"
	accountIdPathPrefix := "/restapi/v2.1/accounts/" + accountId

	since, err := time.Parse(time.RFC3339, "2026-03-03T16:33:32Z")
	require.NoError(t, err)

	responseEnvelopesFirstPage := testutils.DataFromFile(t, "read-envelopes-first-page.json")
	responseEnvelopesSecondPage := testutils.DataFromFile(t, "read-envelopes-second-page.json")
	responseEnvelopesDateRange := testutils.DataFromFile(t, "read-envelopes-date-range.json")

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
		{
			Name: "Read Envelopes with date range",
			Input: common.ReadParams{
				ObjectName: "envelopes",
				Fields:     connectors.Fields("envelopeId", "documentsUri"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path(accountIdPathPrefix + "/envelopes"),
				Then:  mockserver.Response(http.StatusOK, responseEnvelopesDateRange),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"envelopeid":   "49a523cb-d283-8f6f-8060-2a04b73f01b7",
						"documentsuri": "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/documents",
					},
					Raw: map[string]any{
						"envelopeId":            "49a523cb-d283-8f6f-8060-2a04b73f01b7",
						"documentsUri":          "/envelopes/49a523cb-d283-8f6f-8060-2a04b73f01b7/documents",
						"statusChangedDateTime": "2026-03-04T15:18:13.7170000Z",
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
