package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	dr "github.com/amp-labs/connectors/providers/drift"
	"github.com/amp-labs/connectors/test/drift"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := drift.GetConnector(ctx)

	err := testCreatingContacts(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateContacts(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingNewConversation(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingContacts(ctx context.Context, conn *dr.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"attributes": map[string]any{
				"email": gofakeit.Email(),
			},
		},
	}

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

func testUpdateContacts(ctx context.Context, conn *dr.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "24691647847",
		RecordData: map[string]any{
			"attributes": map[string]any{
				"externalId": gofakeit.RandomString([]string{"abcdeghijklmno"}),
			},
		},
	}

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

func testCreatingNewConversation(ctx context.Context, conn *dr.Connector) error {
	params := common.WriteParams{
		ObjectName: "conversations",
		RecordData: map[string]any{
			"email": "josephkarage@email.com",
			"message": map[string]any{
				"body": "A conversation was started <a href='www.withampersand.com'>here</a>, let's resume from drift!",
				"attributes": map[string]any{
					"integrationSource": "Message from facebook",
				},
			},
		},
	}

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
