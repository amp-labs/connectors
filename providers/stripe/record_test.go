package stripe

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// GetRecordsByIdsInput represents the input parameters for GetRecordsByIds method.
type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

type GetRecordsByIdsTestCase = testroutines.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	paymentIntentWithAssociations := testutils.DataFromFile(t, "read/payment_intents/expand_payment_method_charge.json")

	tests := []GetRecordsByIdsTestCase{
		{
			Name: "Empty IDs returns empty slice",
			Input: GetRecordsByIdsInput{
				ObjectName:   "payment_intents",
				Ids:          []string{},
				Fields:       []string{"id", "amount", "currency"},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     []common.ReadResultRow{},
			ExpectedErrs: nil,
		},
		{
			Name: "Missing object name",
			Input: GetRecordsByIdsInput{
				ObjectName:   "",
				Ids:          []string{"pi_123"},
				Fields:       []string{"id"},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Missing fields",
			Input: GetRecordsByIdsInput{
				ObjectName:   "payment_intents",
				Ids:          []string{"pi_123"},
				Fields:       []string{},
				Associations: nil,
			},
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "With associations (expand)",
			Input: GetRecordsByIdsInput{
				ObjectName:   "payment_intents",
				Ids:          []string{"pi_3SsAwzF6iHem4voo03GfTErP"},
				Fields:       []string{"id", "amount", "currency"},
				Associations: []string{"payment_method", "latest_charge"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/payment_intents/pi_3SsAwzF6iHem4voo03GfTErP"),
					mockcond.QueryParam("expand[]", "payment_method"),
					mockcond.QueryParam("expand[]", "latest_charge"),
				},
				Then: mockserver.Response(http.StatusOK, paymentIntentWithAssociations),
			}.Server(),
			Comparator: compareReadResultRows,
			Expected: []common.ReadResultRow{
				{
					Id: "pi_3SsAwzF6iHem4voo03GfTErP",
					Fields: map[string]any{
						"id":       "pi_3SsAwzF6iHem4voo03GfTErP",
						"amount":   float64(100),
						"currency": "usd",
					},
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

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(
				context.Background(),
				tt.Input.ObjectName,
				tt.Input.Ids,
				tt.Input.Fields,
				tt.Input.Associations,
			)

			tt.Validate(t, err, result)
		})
	}
}

// compareReadResultRows compares two slices of ReadResultRow by wrapping them
// into ReadResult and using existing utilities for Fields and Raw, plus association validation.
func compareReadResultRows(_ string, actual, expected []common.ReadResultRow) bool {
	// Wrap slices into ReadResult to use existing utilities
	actualResult := &common.ReadResult{Data: actual}
	expectedResult := &common.ReadResult{Data: expected}

	// Use existing utilities for Fields and Raw
	if !mockutils.ReadResultComparator.SubsetFields(actualResult, expectedResult) {
		return false
	}

	// Check that Raw is populated for all rows
	for i := range expectedResult.Data {
		if actualResult.Data[i].Raw == nil {
			return false
		}
	}

	// Validate associations if expected
	for i := range expectedResult.Data {
		expectedAssoc := expectedResult.Data[i].Associations
		actualAssoc := actualResult.Data[i].Associations

		// If expected has no associations, actual can have none or some (we don't care)
		if len(expectedAssoc) == 0 {
			continue
		}

		// If expected has associations but actual doesn't, that's a failure
		if len(actualAssoc) == 0 {
			return false
		}

		// Check each expected association type
		for assocType, expectedAssociations := range expectedAssoc {
			actualAssociations, ok := actualAssoc[assocType]
			if !ok || len(actualAssociations) != len(expectedAssociations) {
				return false
			}

			for j, expectedAssoc := range expectedAssociations {
				actualAssoc := actualAssociations[j]

				// Check ObjectId matches
				if expectedAssoc.ObjectId != "" && actualAssoc.ObjectId != expectedAssoc.ObjectId {
					return false
				}

				// Verify Raw is populated (contains the full associated object)
				if actualAssoc.Raw == nil {
					return false
				}

				// If expected Raw is specified, verify key fields match (subset matching)
				if expectedAssoc.Raw != nil {
					for key, expectedVal := range expectedAssoc.Raw {
						actualVal, exists := actualAssoc.Raw[key]
						if !exists || !reflect.DeepEqual(actualVal, expectedVal) {
							return false
						}
					}
				}
			}
		}
	}

	return true
}
