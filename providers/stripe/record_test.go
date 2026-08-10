package stripe

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	paymentIntentWithAssociations := testutils.DataFromFile(t, "read/payment_intents/expand_payment_method_charge.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name: "Empty IDs returns empty slice",
			Input: testconn.ReadByIdsParams{
				ObjectName:   "payment_intents",
				RecordIds:    []string{},
				Fields:       []string{"id", "amount", "currency"},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     []common.ReadResultRow{},
			ExpectedErrs: nil,
		},
		{
			Name: "Missing object name",
			Input: testconn.ReadByIdsParams{
				ObjectName:   "",
				RecordIds:    []string{"pi_123"},
				Fields:       []string{"id"},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Missing fields",
			Input: testconn.ReadByIdsParams{
				ObjectName:   "payment_intents",
				RecordIds:    []string{"pi_123"},
				Fields:       []string{},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "With associations (expand)",
			Input: testconn.ReadByIdsParams{
				ObjectName:   "payment_intents",
				RecordIds:    []string{"pi_3SsAwzF6iHem4voo03GfTErP"},
				Fields:       []string{"id", "amount", "currency"},
				Associations: []string{"payment_method", "latest_charge"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/payment_intents/pi_3SsAwzF6iHem4voo03GfTErP"),
					mockcond.QueryParam("expand[]", "payment_method", "latest_charge"),
				},
				Then: mockserver.Response(http.StatusOK, paymentIntentWithAssociations),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{
				{
					Id: "pi_3SsAwzF6iHem4voo03GfTErP",
					Fields: map[string]any{
						"id":       "pi_3SsAwzF6iHem4voo03GfTErP",
						"amount":   float64(100),
						"currency": "usd",
					},
					Raw: map[string]any{"currency": "usd"},
					Associations: map[string][]common.Association{
						"payment_method": {
							{
								ObjectId: "pm_1SsAwzF6iHem4voorZHZEMKc",
								Raw:      nil, // raw will contain full payment_method object, just verify ObjectId
							},
						},
						"latest_charge": {
							{
								ObjectId: "ch_3SsAwzF6iHem4voo0kHo2r3C",
								Raw:      nil,
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
