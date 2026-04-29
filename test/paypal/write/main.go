package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/paypal"
	paypalTest "github.com/amp-labs/connectors/test/paypal"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := paypalTest.GetPayPalConnector(ctx)

	if err := run(ctx, conn); err != nil {
		utils.Fail("write test failed", "error", err)
	}
}

func run(ctx context.Context, conn *paypal.Connector) error {
	// Products: create then patch-update.
	productID, err := testCreatingProduct(ctx, conn)
	if err != nil {
		return err
	}

	if err = testUpdatingProduct(ctx, conn, productID); err != nil {
		return err
	}

	// Plans: create (requires a product) then patch-update.
	planID, err := testCreatingPlan(ctx, conn, productID)
	if err != nil {
		return err
	}

	if err = testUpdatingPlan(ctx, conn, planID); err != nil {
		return err
	}

	// Invoices: create then full-replace update (PUT).
	invoiceID, err := testCreatingInvoice(ctx, conn)
	if err != nil {
		return err
	}

	if err = testUpdatingInvoice(ctx, conn, invoiceID); err != nil {
		return err
	}

	// Orders: create only (update requires buyer approval flow).
	if err = testCreatingOrder(ctx, conn); err != nil {
		return err
	}

	return nil
}

func testCreatingProduct(ctx context.Context, conn *paypal.Connector) (string, error) {
	slog.Info("Creating product...")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"name":        "Test Product",
			"description": "Created by write test",
			"type":        "SERVICE",
			"category":    "SOFTWARE",
		},
	})
	if err != nil {
		return "", err
	}

	slog.Info("Created product", "id", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdatingProduct(ctx context.Context, conn *paypal.Connector, productID string) error {
	slog.Info("Updating product...", "id", productID)

	// PayPal products use JSON Patch format for updates.
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordId:   productID,
		RecordData: []map[string]any{
			{"op": "replace", "path": "/description", "value": "Updated by write test"},
		},
	})
	if err != nil {
		return err
	}

	slog.Info("Updated product", "id", productID)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreatingPlan(ctx context.Context, conn *paypal.Connector, productID string) (string, error) {
	slog.Info("Creating plan...", "product_id", productID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "plans",
		RecordData: map[string]any{
			"product_id": productID,
			"name":       "Test Plan",
			"billing_cycles": []map[string]any{
				{
					"frequency":      map[string]any{"interval_unit": "MONTH", "interval_count": 1},
					"tenure_type":    "REGULAR",
					"sequence":       1,
					"total_cycles":   12,
					"pricing_scheme": map[string]any{"fixed_price": map[string]any{"value": "9.99", "currency_code": "USD"}},
				},
			},
			"payment_preferences": map[string]any{
				"auto_bill_outstanding":     true,
				"setup_fee_failure_action":  "CONTINUE",
				"payment_failure_threshold": 3,
			},
		},
	})
	if err != nil {
		return "", err
	}

	slog.Info("Created plan", "id", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdatingPlan(ctx context.Context, conn *paypal.Connector, planID string) error {
	slog.Info("Updating plan...", "id", planID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "plans",
		RecordId:   planID,
		RecordData: []map[string]any{
			{"op": "replace", "path": "/name", "value": "Updated Test Plan"},
		},
	})
	if err != nil {
		return err
	}

	slog.Info("Updated plan", "id", planID)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreatingInvoice(ctx context.Context, conn *paypal.Connector) (string, error) {
	slog.Info("Creating invoice...")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "invoices",
		RecordData: map[string]any{
			"detail": map[string]any{
				"currency_code": "USD",
				"note":          "Created by write test",
			},
			"items": []map[string]any{
				{
					"name":        "Consulting",
					"unit_amount": map[string]any{"currency_code": "USD", "value": "100.00"},
					"quantity":    "1",
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	slog.Info("Created invoice", "id", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdatingInvoice(ctx context.Context, conn *paypal.Connector, invoiceID string) error {
	slog.Info("Updating invoice...", "id", invoiceID)

	// Invoice update uses PUT (full replacement).
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "invoices",
		RecordId:   invoiceID,
		RecordData: map[string]any{
			"detail": map[string]any{
				"currency_code": "USD",
				"note":          "Updated by write test",
			},
			"items": []map[string]any{
				{
					"name":        "Consulting",
					"unit_amount": map[string]any{"currency_code": "USD", "value": "150.00"},
					"quantity":    "1",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	slog.Info("Updated invoice", "id", invoiceID)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreatingOrder(ctx context.Context, conn *paypal.Connector) error {
	slog.Info("Creating order...")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "orders",
		RecordData: map[string]any{
			"intent": "CAPTURE",
			"purchase_units": []map[string]any{
				{
					"amount": map[string]any{
						"currency_code": "USD",
						"value":         "10.00",
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	slog.Info("Created order", "id", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
