package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/outplay"
	"github.com/amp-labs/connectors/test/outplay"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := outplay.GetOutplayConnector(ctx)

	err := testCreatingPropects(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingProspectAccounts(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingPropects(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "prospect",
		RecordData: map[string]any{
			"emailid":   gofakeit.Email(),
			"firstname": gofakeit.FirstName(),
			"lastname":  gofakeit.LastName(),
		},
	}

	slog.Info("Creating prospects...")

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

func testCreatingProspectAccounts(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "prospectaccount",
		RecordData: map[string]any{
			"name":        gofakeit.Company(),
			"externalid":  gofakeit.UUID(),
			"description": "account description",
		},
	}

	slog.Info("Creating prospect accounts...")

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
