package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/teamleader"
	"github.com/amp-labs/connectors/test/teamleader"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := teamleader.GetConnector(ctx)

	id, err := testCreatingContacts(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateContact(ctx, conn, id)
	if err != nil {
		return err
	}

	err = testCreateDeals(ctx, conn, id)
	if err != nil {
		return err
	}

	err = testCreateDealPipelines(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingContacts(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      gofakeit.Email(),
		},
	}

	slog.Info("Creating contact...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["id"].(string), nil
}

func testUpdateContact(ctx context.Context, conn *cc.Connector, contactID string) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: map[string]any{
			"id":         contactID,
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      gofakeit.Email(),
		},
	}

	slog.Info("Updating contact...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreateDeals(ctx context.Context, conn *cc.Connector, customerId string) error {
	params := common.WriteParams{
		ObjectName: "deals",
		RecordData: map[string]any{
			"title": gofakeit.BeerName(),
			"lead": map[string]any{
				"customer": map[string]any{
					"type": "contact",
					"id":   customerId, // Use the contact ID created earlier
				},
			},
		},
	}

	slog.Info("Creating deal...")

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

func testCreateDealPipelines(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "dealPipelines",
		RecordData: map[string]any{
			"name": gofakeit.BeerName(),
		},
	}

	slog.Info("Creating deal pipeline...")

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
