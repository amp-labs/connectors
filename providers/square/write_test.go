package square

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createCustomerResponse := testutils.DataFromFile(t, "write/create-customer.json")
	createGiftCardResponse := testutils.DataFromFile(t, "write/create-gift-card.json")
	createPaymentResponse := testutils.DataFromFile(t, "write/create-payment.json")
	updateBookingResponse := testutils.DataFromFile(t, "write/update-booking.json")
	upsertCatalogResponse := testutils.DataFromFile(t, "write/upsert-catalog-object.json")
	createDefinitionResponse := testutils.DataFromFile(t, "write/create-custom-attribute-definition.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Read-only object is not supported",
			Input: common.WriteParams{
				ObjectName: "payouts",
				RecordData: map[string]any{"foo": "bar"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			// Customer creation takes a flat body while the response wraps the
			// record in a "customer" envelope.
			Name: "Create customer with flat request body",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"given_name":    "Amelia",
					"family_name":   "Earhart",
					"email_address": "amelia.earhart@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/customers"),
					mockcond.Body(`{
						"given_name": "Amelia",
						"family_name": "Earhart",
						"email_address": "amelia.earhart@example.com"
					}`),
				},
				Then: mockserver.Response(http.StatusOK, createCustomerResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "JDKYHBWT1D4F8MFH63DBMEN8Y4",
				Data: map[string]any{
					"id":         "JDKYHBWT1D4F8MFH63DBMEN8Y4",
					"given_name": "Amelia",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// The record is wrapped under "gift_card", location_id is hoisted to
			// the top level, and a missing idempotency_key is auto-generated.
			Name: "Create gift card wraps record and injects idempotency key",
			Input: common.WriteParams{
				ObjectName: "gift_cards",
				RecordData: map[string]any{
					"type":        "DIGITAL",
					"location_id": "L72WA3EVDJ8FY",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/gift-cards"),
					mockcond.Body(`{
						"location_id": "L72WA3EVDJ8FY",
						"gift_card": {"type": "DIGITAL"}
					}`, mockcond.IgnoreBodyField("idempotency_key")),
				},
				Then: mockserver.Response(http.StatusOK, createGiftCardResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gftc:6d55a92ac6db4f4c9b78be6a9b1a1a1a",
				Data: map[string]any{
					"id":    "gftc:6d55a92ac6db4f4c9b78be6a9b1a1a1a",
					"state": "PENDING",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// A caller-supplied idempotency_key is passed through untouched.
			Name: "Create payment passes caller idempotency key through",
			Input: common.WriteParams{
				ObjectName: "payments",
				RecordData: map[string]any{
					"idempotency_key": "my-key-123",
					"source_id":       "cnon:card-nonce-ok",
					"amount_money":    map[string]any{"amount": 1000, "currency": "USD"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/payments"),
					mockcond.Body(`{
						"idempotency_key": "my-key-123",
						"source_id": "cnon:card-nonce-ok",
						"amount_money": {"amount": 1000, "currency": "USD"}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, createPaymentResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "bP9mAsEMYPUGjjGNaNO5ZDVyLhSZY",
				Data: map[string]any{
					"id":     "bP9mAsEMYPUGjjGNaNO5ZDVyLhSZY",
					"status": "COMPLETED",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update booking puts wrapped record to id path",
			Input: common.WriteParams{
				ObjectName: "bookings",
				RecordId:   "zkbugu5jyvgfdqrp",
				RecordData: map[string]any{
					"version":  1,
					"start_at": "2026-08-01T15:00:00Z",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/bookings/zkbugu5jyvgfdqrp"),
					mockcond.Body(`{
						"booking": {"version": 1, "start_at": "2026-08-01T15:00:00Z"}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, updateBookingResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "zkbugu5jyvgfdqrp",
				Data: map[string]any{
					"id":     "zkbugu5jyvgfdqrp",
					"status": "ACCEPTED",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// Catalog updates go through the upsert endpoint: POST to the same
			// path with the record id and current version inside the object.
			Name: "Update catalog object via upsert",
			Input: common.WriteParams{
				ObjectName: "catalog",
				RecordId:   "W62UWFY35CWMYGVWK6TWJPNI",
				RecordData: map[string]any{
					"id":      "W62UWFY35CWMYGVWK6TWJPNI",
					"type":    "ITEM",
					"version": 2,
					"item_data": map[string]any{
						"name": "Cocoa",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/catalog/object"),
					mockcond.Body(`{
						"object": {
							"id": "W62UWFY35CWMYGVWK6TWJPNI",
							"type": "ITEM",
							"version": 2,
							"item_data": {"name": "Cocoa"}
						}
					}`, mockcond.IgnoreBodyField("idempotency_key")),
				},
				Then: mockserver.Response(http.StatusOK, upsertCatalogResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "W62UWFY35CWMYGVWK6TWJPNI",
				Data: map[string]any{
					"id":      "W62UWFY35CWMYGVWK6TWJPNI",
					"version": float64(3),
				},
			},
			ExpectedErrs: nil,
		},
		{
			// Custom attribute definitions identify records by "key", not "id".
			Name: "Create customer custom attribute definition uses key as id",
			Input: common.WriteParams{
				ObjectName: "customers/custom_attribute_definitions",
				RecordData: map[string]any{
					"key":  "favorite-drink",
					"name": "Favorite Drink",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/customers/custom-attribute-definitions"),
					mockcond.Body(`{
						"custom_attribute_definition": {"key": "favorite-drink", "name": "Favorite Drink"}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, createDefinitionResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "favorite-drink",
				Data: map[string]any{
					"key":  "favorite-drink",
					"name": "Favorite Drink",
				},
			},
			ExpectedErrs: nil,
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
