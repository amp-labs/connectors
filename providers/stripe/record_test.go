package stripe

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen
	t.Parallel()

	paymentIntentWithAssociations := testutils.DataFromFile(t, "read/payment_intents/expand_payment_method_charge.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name: "Empty IDs returns empty slice",
			Input: testconn.ReadByIdsParams{
				ObjectName: "payment_intents",
				RecordIds:  []string{},
				Fields:     []string{"id", "amount", "currency"},
			},
			Server:       mockserver.Dummy(),
			Expected:     []common.ReadResultRow{},
			ExpectedErrs: nil,
		},
		{
			Name: "Missing object name",
			Input: testconn.ReadByIdsParams{
				ObjectName: "",
				RecordIds:  []string{"pi_123"},
				Fields:     []string{"id"},
			},
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Missing fields",
			Input: testconn.ReadByIdsParams{
				ObjectName: "payment_intents",
				RecordIds:  []string{"pi_123"},
				Fields:     []string{},
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

// compareReadResultRows compares two slices of ReadResultRow by wrapping them
// into ReadResult and using existing utilities for Fields and Raw, plus association validation.
func compareReadResultRows(_ string, actual, expected []common.ReadResultRow) *testutils.CompareResult {
	// Wrap slices into ReadResult to use existing utilities
	actualResult := &common.ReadResult{Data: actual}
	expectedResult := &common.ReadResult{Data: expected}

	// Use existing utilities for Fields and Raw
	result := mockutils.ReadResultComparator.SubsetFields(actualResult, expectedResult)

	for i := range expectedResult.Data {
		// Check that Raw is populated for all rows
		if actualResult.Data[i].Raw == nil {
			result.AddDiff("row [%v] has no Raw data", i)
		}

		// Validate associations if expected
		result.Merge(compareRowAssociations(i, actualResult.Data[i].Associations, expectedResult.Data[i].Associations))
	}

	return result
}

func compareRowAssociations( // nolint:cyclop
	rowIndex int,
	actualAssoc, expectedAssoc map[string][]common.Association,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	// If expected has no associations, actual can have none or some (we don't care)
	if len(expectedAssoc) == 0 {
		return result
	}

	// If expected has associations but actual doesn't, that's a failure
	if len(actualAssoc) == 0 {
		return result.AddDiff("row [%v] has no associations", rowIndex)
	}

	// Check each expected association type
	for assocType, expectedAssociations := range expectedAssoc {
		actualAssociations, ok := actualAssoc[assocType]
		if !ok || len(actualAssociations) != len(expectedAssociations) {
			result.AddDiff("row [%v] association [%v] is missing or has mismatching length", rowIndex, assocType)

			continue
		}

		for j, expectedItem := range expectedAssociations {
			actualItem := actualAssociations[j]

			// Check ObjectId matches
			if expectedItem.ObjectId != "" {
				result.Assert("association ObjectId", expectedItem.ObjectId, actualItem.ObjectId)
			}

			// Verify Raw is populated (contains the full associated object)
			if actualItem.Raw == nil {
				result.AddDiff("row [%v] association [%v][%v] has no Raw data", rowIndex, assocType, j)
			}

			// If expected Raw is specified, verify key fields match (subset matching)
			for key, expectedVal := range expectedItem.Raw {
				actualVal, exists := actualItem.Raw[key]
				if !exists || !reflect.DeepEqual(actualVal, expectedVal) {
					result.AddDiff("row [%v] association [%v][%v] Raw key [%v] mismatch", rowIndex, assocType, j, key)
				}
			}
		}
	}

	return result
}
