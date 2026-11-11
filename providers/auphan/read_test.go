package auphan

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responsePayments := testutils.DataFromFile(t, "payments.json")
	responsePaymentsSingleRows := testutils.DataFromFile(t, "payments-single-rows.json")
	responsePaymentsEmpty := testutils.DataFromFile(t, "payments-empty.json")
	responseValidationError := testutils.DataFromFile(t, "error-validation.json")
	responseInvalidToken := testutils.DataFromFile(t, "error-invalid-token.json")
	responseStoresFirstPage := testutils.DataFromFile(t, "stores-page-1.json")
	responseStoresLastPage := testutils.DataFromFile(t, "stores-page-2.json")
	responsePettyCash := testutils.DataFromFile(t, "petty-cash.json")
	responseTaxDetailRows := testutils.DataFromFile(t, "tax-detail-rows.json")

	// A day in a fixed zone that is not UTC: request times must keep it, unconverted.
	zone := time.FixedZone("store", -5*60*60)
	since := time.Date(2026, 11, 5, 0, 0, 0, 0, zone)
	until := time.Date(2026, 11, 5, 23, 59, 59, 0, zone)

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "orders", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Payments summary rows are merged in pairs",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment_type", "payment", "tips", "count"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v1.php/reports/execute"),
					mockcond.Body(`{
						"store_id": 73921,
						"params": {
							"type": "payments",
							"columns": ["count", "payment", "payment_type", "tips"],
							"filters": {
								"selection": "close_date",
								"status": [2],
								"remove_cash_change": 1,
								"net_cash_deducts_tips": 1,
								"date_from": "2026-11-05 00:00:00",
								"date_to": "2026-11-05 23:59:59"
							}
						}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responsePayments),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"payment_type": "MC",
						"payment":      "1411.96",
						"tips":         "180.76",
						"count":        "36",
					},
					Raw: map[string]any{
						"date":         "Nov 05, 2026",
						"payment_type": "MC",
						"gl_code":      "",
						"payment":      "1411.96",
						"tips":         "180.76",
						"count":        "36",
					},
				}, {
					Fields: map[string]any{
						"payment_type": "Cash",
						"payment":      "490.46",
						"tips":         nil,
						"count":        "21",
					},
					Raw: map[string]any{
						"date":         "Nov 05, 2026",
						"payment_type": "Cash",
						"gl_code":      "",
						"payment":      "490.46",
						"tips":         nil,
						"count":        "21",
					},
				}},
				Done: true,
			},
		},
		{
			Name: "Petty cash is read without a date selection, which that report rejects",
			Input: common.ReadParams{
				ObjectName: "petty_cash",
				Fields:     connectors.Fields("petty_cash_type", "reason", "tracking_date", "amount"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{
					"store_id": 73921,
					"params": {
						"type": "petty_cash",
						"columns": ["amount", "petty_cash_type", "reason", "tracking_date"],
						"filters": {"date_from": "2026-11-05 00:00:00", "date_to": "2026-11-05 23:59:59"}
					}
				}`),
				Then: mockserver.Response(http.StatusOK, responsePettyCash),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"petty_cash_type": "Petty Cash Out",
						"amount":          "-80.26",
					},
					Raw: map[string]any{"reason": "Alex tips"},
				}, {
					Fields: map[string]any{
						"petty_cash_type": "Petty Cash In",
						"amount":          "11.50",
					},
					Raw: map[string]any{"tracking_date": "Nov 05, 26 11:13 pm"},
				}},
				Done: true,
			},
		},
		{
			Name: "Every measure row under a dimension row gets that group's dimensions",
			Input: common.ReadParams{
				ObjectName: "tax",
				Fields:     connectors.Fields("name", "tax_date", "tax_total"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTaxDetailRows),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 4,
				Data: []common.ReadResultRow{
					taxRow("Delivery Fee", "0.00"),
					taxRow("Sales Tax", "1.46"),
					taxRow("Sales Tax", "1.58"),
					taxRow("Sales Tax", "5.12"),
				},
				Done: true,
			},
		},
		{
			Name: "Report rows that are already whole are kept as they are",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment_type", "payment", "count"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responsePaymentsSingleRows),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"payment_type": "Visa", "payment": "4680.61", "count": "153"},
					Raw:    map[string]any{"payment_type": "Visa", "payment": "4680.61", "count": "153"},
				}, {
					Fields: map[string]any{"payment_type": "Amex", "payment": "232.70", "count": "6"},
					Raw:    map[string]any{"payment_type": "Amex", "payment": "232.70", "count": "6"},
				}},
				Done: true,
			},
		},
		{
			Name: "Report with no rows",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment_type", "payment"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responsePaymentsEmpty),
			}.Server(),
			Expected: &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
		},
		{
			Name: "success false with HTTP 200 is a bad request",
			Input: common.ReadParams{
				ObjectName: "sales",
				Fields:     connectors.Fields("net_sales"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseValidationError),
			}.Server(),
			ExpectedErrs: []error{common.ErrBadRequest, testutils.StringError("Invalid selection")},
		},
		{
			Name: "Invalid token with HTTP 200 is an access token error",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseInvalidToken),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name: "Invalid token with HTTP 401 is an access token error",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseInvalidToken),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken, testutils.StringError("Token not recognized")},
		},
		{
			Name: "Gateway text 403 for an unknown token is an access token error",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentText(),
				Always: mockserver.Response(http.StatusForbidden, []byte("Forbidden")),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name: "JSON 403 from the store stays forbidden",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, []byte(
					`{"success": false, "error": {"code": "MODULE_ACCESS_DENIED", "message": "Read denied"}}`)),
			}.Server(),
			ExpectedErrs: []error{common.ErrForbidden, testutils.StringError("Read denied")},
		},
		{
			Name: "Rate limit is reported",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("payment"),
				Since:      since,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusTooManyRequests),
			}.Server(),
			ExpectedErrs: []error{common.ErrLimitExceeded},
		},
		{
			Name:  "Stores first page points to the next page",
			Input: common.ReadParams{ObjectName: "stores", Fields: connectors.Fields("store_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v1.php/settings/stores.list"),
					mockcond.Body(`{"store_id": 73921, "params": {"page": 1, "per_page": 200}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseStoresFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"store_id": "73921"},
					Raw:    map[string]any{"store_id": "73921"},
				}},
				NextPage: "2",
				Done:     false,
			},
		},
		{
			Name: "Stores last page is done",
			Input: common.ReadParams{
				ObjectName: "stores",
				Fields:     connectors.Fields("store_id"),
				NextPage:   "2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Body(`{"store_id": 73921, "params": {"page": 2, "per_page": 200}}`),
				Then:  mockserver.Response(http.StatusOK, responseStoresLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"store_id": "73922"},
					Raw:    map[string]any{"store_id": "73922"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	return constructTestConnectorWithMetadata(serverURL, nil)
}

func constructTestConnectorWithMetadata(serverURL string, metadata map[string]string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "73921",
		Metadata:            metadata,
	})
	if err != nil {
		return nil, err
	}

	// Send calls to the mock server, keeping the /api/v1.php path.
	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}

func TestSandboxRead(t *testing.T) {
	t.Parallel()

	responseStores := testutils.DataFromFile(t, "stores-page-2.json")

	tests := []testconn.TestCaseRead{
		{
			Name:  "Sandbox puts the demo name in the path and the metadata store ID in the body",
			Input: common.ReadParams{ObjectName: "stores", Fields: connectors.Fields("store_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/AmpersandDemo/api/v1.php/settings/stores.list"),
					mockcond.Body(`{"store_id": 1001, "params": {"page": 1, "per_page": 200}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseStores),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"store_id": "73922"},
					Raw:    map[string]any{"store_id": "73922"},
				}},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				connector, err := NewSandboxConnector(common.ConnectorParams{
					AuthenticatedClient: mockutils.NewClient(),
					Workspace:           "AmpersandDemo",
					Metadata:            map[string]string{"storeId": "1001"},
				})
				if err != nil {
					return nil, err
				}

				connector.SetUnitTestMockServerBaseUrl(tt.Server.URL)

				return connector, nil
			})
		})
	}
}

func TestSandboxNeedsStoreID(t *testing.T) {
	t.Parallel()

	_, err := NewSandboxConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "AmpersandDemo",
	})
	if !errors.Is(err, errInvalidStoreID) {
		t.Fatalf("expected errInvalidStoreID, got %v", err)
	}
}

func taxRow(name, total string) common.ReadResultRow {
	fields := map[string]any{"name": name, "tax_date": "Nov 05, 2026", "tax_total": total}

	return common.ReadResultRow{Fields: fields, Raw: fields}
}
