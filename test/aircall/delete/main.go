package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aircall"
	testAircall "github.com/amp-labs/connectors/test/aircall"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testAircall.GetAircallConnector(ctx)

	slog.Info("Testing delete operations for contacts")

	// Step 1: Create a contact to delete
	// We use a unique index to ensure unique phone numbers/emails
	index := 999
	slog.Info("Step 1: Creating a contact to delete")
	recordID, err := testCreateContact(ctx, conn, index)
	if err != nil {
		slog.Error("Failed to create contact", "error", err)
		return
	}

	// Step 2: Delete the contact
	slog.Info("Step 2: Deleting the contact", "recordId", recordID)
	if err := testDeleteContact(ctx, conn, recordID); err != nil {
		slog.Error("Failed to delete contact", "error", err)
		return
	}

	slog.Info("✅ Contact deleted successfully!")
}

func testCreateContact(ctx context.Context, conn *aircall.Connector, index int) (string, error) {
	// Create realistic contact data using gofakeit
	recordData := map[string]any{
		"first_name":  gofakeit.FirstName(),
		"last_name":   gofakeit.LastName(),
		"information": fmt.Sprintf("Test delete contact #%d - created for deletion test", index),
		"phone_numbers": []map[string]any{
			{
				"label": "Work",
				"value": fmt.Sprintf("+1%d", gofakeit.Number(2000000000, 8999999999)),
			},
		},
		"emails": []map[string]any{
			{
				"label": "Work",
				"value": gofakeit.Email(),
			},
		},
	}

	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating contact: %w", err)
	}

	slog.Info("Contact created successfully", "recordId", res.RecordId)
	return res.RecordId, nil
}

func testDeleteContact(ctx context.Context, conn *aircall.Connector, recordID string) error {
	params := common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   recordID,
	}

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("error deleting contact: %w", err)
	}

	slog.Info("Delete operation successful", "success", res.Success)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	return nil
}
