package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/chargebee"
	"github.com/amp-labs/connectors/test/chargebee"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := chargebee.GetChargebeeConnector(ctx)

	id, err := testCreatingCustomers(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdatingCustomers(ctx, conn, id); err != nil {
		return err
	}

	_, err = testCreatingPromotionalCredits(ctx, conn, id)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCustomers(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"first_name":        "John",
			"last_name":         "Doe",
			"email":             gofakeit.Email(),
			"cf_customer_hobby": gofakeit.Hobby(),
		},
	}

	slog.Info("Creating customers...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["id"].(string), nil
}

func testUpdatingCustomers(ctx context.Context, conn *cc.Connector, customerID string) error {
	params := common.WriteParams{
		ObjectName: "customers",
		RecordId:   customerID,
		RecordData: map[string]any{
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      gofakeit.Email(),
		},
	}

	slog.Info("Updating customers...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreatingPromotionalCredits(ctx context.Context, conn *cc.Connector, customerId string) (string, error) {
	params := common.WriteParams{
		ObjectName: "promotional_credits",
		RecordData: map[string]any{
			"amount":      1000,
			"customer_id": customerId,
			"description": gofakeit.Sentence(10),
		},
	}

	slog.Info("Creating promotional credits...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["id"].(string), nil
}
