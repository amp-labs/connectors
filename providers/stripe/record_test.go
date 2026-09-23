package stripe

import (
	"net/http"
	"reflect"
	"testing"
	"time"

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
			// Webhook events carry singular object types (SubscriptionEvent.ObjectName
			// returns e.g. "payment_intent"); the fetch URL must use the plural collection.
			Name: "Singular object name resolves to plural collection",
			Input: testconn.ReadByIdsParams{
				ObjectName: "payment_intent",
				RecordIds:  []string{"pi_3SsAwzF6iHem4voo03GfTErP"},
				Fields:     []string{"id", "amount", "currency"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/payment_intents/pi_3SsAwzF6iHem4voo03GfTErP"),
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
				},
			},
			ExpectedErrs: nil,
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

func TestExtractAssociationsDotPath(t *testing.T) {
	t.Parallel()

	record := map[string]any{
		"id":     "pi_123",
		"object": "payment_intent",
		"customer": map[string]any{
			"id":     "cus_123",
			"object": "customer",
			"default_source": map[string]any{
				"id":     "card_123",
				"object": "card",
				"last4":  "4242",
			},
		},
	}

	associations := extractAssociations(record, []string{"customer", "customer.default_source", "customer.missing"})

	result := testutils.NewCompareResult()

	customer, found := associations["customer"]
	if result.Assert("customer association found", true, found) {
		result.Assert("customer ObjectId", "cus_123", customer[0].ObjectId)
	}

	defaultSource, found := associations["customer.default_source"]
	if result.Assert("customer.default_source association found", true, found) {
		result.Assert("customer.default_source ObjectId", "card_123", defaultSource[0].ObjectId)
	}

	_, found = associations["customer.missing"]
	result.Assert("customer.missing association absent", false, found)

	result.Validate(t, "extract associations by dot-separated path")
}

// Fetches now run concurrently, so the order rows come back in must come from
// the order the ids were requested, not from whichever response lands first.
// The server answers the first id slowly to make an order-of-completion bug
// deterministic rather than a flake.
func TestGetRecordsByIds_PreservesRequestedOrder(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.Method(http.MethodGet),
					mockcond.Path("/v1/payment_intents/pi_first"),
				},
				Then: func(w http.ResponseWriter, _ *http.Request) {
					time.Sleep(50 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":"pi_first","amount":100}`))
				},
			},
			{
				If: mockcond.And{
					mockcond.Method(http.MethodGet),
					mockcond.Path("/v1/payment_intents/pi_second"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"pi_second","amount":200}`),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	batchRes, err := conn.GetRecordsByIds(t.Context(), "payment_intents",
		[]string{"pi_first", "pi_second"}, []string{"id"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	rows := batchRes.Rows
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].Id != "pi_first" || rows[1].Id != "pi_second" {
		t.Fatalf("rows out of requested order: got %q, %q", rows[0].Id, rows[1].Id)
	}
}

// A skipped not-found id must not leave a gap or shift the surviving rows.
func TestGetRecordsByIds_NotFoundSkippedKeepsOrder(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.Method(http.MethodGet),
					mockcond.Path("/v1/payment_intents/pi_first"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"pi_first","amount":100}`),
			},
			{
				If: mockcond.And{
					mockcond.Method(http.MethodGet),
					mockcond.Path("/v1/payment_intents/pi_missing"),
				},
				Then: mockserver.ResponseString(http.StatusNotFound,
					`{"error":{"type":"invalid_request_error","message":"No such payment_intent"}}`),
			},
			{
				If: mockcond.And{
					mockcond.Method(http.MethodGet),
					mockcond.Path("/v1/payment_intents/pi_last"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"pi_last","amount":300}`),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	batchRes, err := conn.GetRecordsByIds(t.Context(), "payment_intents",
		[]string{"pi_first", "pi_missing", "pi_last"}, []string{"id"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	rows := batchRes.Rows
	if len(rows) != 2 {
		t.Fatalf("expected the missing id to be skipped, got %d rows", len(rows))
	}

	if rows[0].Id != "pi_first" || rows[1].Id != "pi_last" {
		t.Fatalf("surviving rows out of order: got %q, %q", rows[0].Id, rows[1].Id)
	}
}
