package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/dropboxsign"
	"github.com/amp-labs/connectors/test/dropboxsign"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := dropboxsign.GetDropboxSignConnector(ctx)

	_, err := testCreatingAccount(ctx, conn)
	if err != nil {
		return err
	}

	apiAppId, err := testCreateApiApp(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateApiApp(ctx, conn, apiAppId)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAccount(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"email_address": gofakeit.Email(),
		},
	}

	slog.Info("Creating account...")

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

	return res.RecordId, nil
}

func testCreateApiApp(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "api_app",
		RecordData: map[string]any{
			"name": gofakeit.Company(),
			"domains": []string{
				gofakeit.DomainName(),
			},
		},
	}

	slog.Info("Creating API App...")

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

	return res.RecordId, nil
}

func testUpdateApiApp(ctx context.Context, conn *cc.Connector, recordId string) error {
	params := common.WriteParams{
		ObjectName: "api_app",
		RecordId:   recordId,
		RecordData: map[string]any{
			"name": "Updated " + gofakeit.Company(),
			"domains": []string{
				gofakeit.DomainName(),
			},
		},
	}

	slog.Info("Updating API App...")

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
