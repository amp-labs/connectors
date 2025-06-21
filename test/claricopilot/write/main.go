package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/claricopilot"
	"github.com/amp-labs/connectors/test/claricopilot"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := claricopilot.GetConnector(ctx)

	id, err := testCreatingContacts(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateContacts(ctx, conn, id)
	if err != nil {
		return err
	}

	err = testCreatingCalls(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateDeals(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateAccounts(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCalls(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "calls",
		RecordData: map[string]any{
			"source_id": gofakeit.UUID(),
			"title":     "Test Call",
			"type":      "RECORDING",
			"status":    "GOTO_MEETING",
			"call_time": "2025-06-05T10:00:00Z",
			"user_emails": []string{
				"integration.user+clari@withampersand.com",
			},
			"source_user_ids": []string{
				"e04483dd-fb82-460a-a14e-b6f3e6a2b7a4",
			},
			"audio_url": "http://file-examples.com/wp-content/storage/2017/11/file_example_MP3_700KB.mp3",
		},
	}

	slog.Info("Creating call...")
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

func testCreatingContacts(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"crm_id":         gofakeit.UUID(),
			"account_crm_id": "salesorcefd",
			"first_name":     "John",
			"last_name":      "Run",
			"job_title":      "string",
			"emails": []string{
				"test1@gmail.com",
			},
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

	return res.Data["crm_id"].(string), nil
}

func testCreateDeals(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "deals",
		RecordData: map[string]any{
			"crm_id":         "12345",
			"amount":         "string",
			"close_date":     "2025-05-24T14:15:22Z",
			"created_date":   "2025-05-24T14:15:22Z",
			"last_contacted": "2025-06-20T14:15:22Z",
		},
	}

	slog.Info("Creating deal...")
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

func testCreateAccounts(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "accounts",
		RecordData: map[string]any{
			"crm_id":      gofakeit.UUID(),
			"name":        gofakeit.Name(),
			"description": gofakeit.Sentence(5),
			"website":     gofakeit.URL(),
			"address":     gofakeit.Address().Address,
			"city":        gofakeit.City(),
			"state":       gofakeit.State(),
			"country":     gofakeit.Country(),
			"zip":         gofakeit.Zip(),
		},
	}

	slog.Info("Creating account...")
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

func testUpdateContacts(ctx context.Context, conn *cc.Connector, id string) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   id,
		RecordData: map[string]any{
			"crm_id":     id,
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
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
