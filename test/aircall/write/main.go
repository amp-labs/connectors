package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	slog.Info("Testing write operations for contacts - Creating 3 test contacts")

	// Create and update 3 contacts
	for i := 1; i <= 3; i++ {
		slog.Info(fmt.Sprintf("=== Test Contact %d ===", i))

		// Test 1: Create a new contact
		slog.Info(fmt.Sprintf("Step 1: Creating contact %d", i))
		recordID, err := testCreateContact(ctx, conn, i)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to create contact %d", i), "error", err)
			continue
		}

		// Test 2: Update the contact
		slog.Info(fmt.Sprintf("Step 2: Updating contact %d", i), "recordId", recordID)
		if err := testUpdateContact(ctx, conn, recordID, i); err != nil {
			slog.Error(fmt.Sprintf("Failed to update contact %d", i), "error", err)
			continue
		}

		slog.Info(fmt.Sprintf("✅ Contact %d completed successfully!", i))
	}

	// Test Tag operations
	slog.Info("=== Testing Tag Operations ===")
	tagID, err := testCreateTag(ctx, conn)
	if err != nil {
		slog.Error("Failed to create tag", "error", err)
	} else {
		if err := testUpdateTag(ctx, conn, tagID); err != nil {
			slog.Error("Failed to update tag", "error", err)
		} else {
			slog.Info("✅ Tag operations completed successfully!")
		}
	}

	slog.Info("All write tests completed successfully!")
}

func testCreateContact(ctx context.Context, conn *aircall.Connector, index int) (string, error) {
	// Create realistic contact data using gofakeit
	recordData := map[string]any{
		"first_name":  gofakeit.FirstName(),
		"last_name":   gofakeit.LastName(),
		"information": fmt.Sprintf("Test contact #%d created at %s", index, time.Now().Format(time.RFC3339)),
		"phone_numbers": []map[string]any{
			{
				"label": "Work",
				"value": gofakeit.Phone(),
			},
		},
		"emails": []map[string]any{
			{
				"label": "Work",
				"value": gofakeit.Email(),
			},
		},
	}

	slog.Info("Creating contact with data", "index", index, "data", recordData)

	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating contact: %w", err)
	}

	slog.Info("Contact created successfully",
		"index", index,
		"recordId", res.RecordId,
		"success", res.Success)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdateContact(ctx context.Context, conn *aircall.Connector, recordID string, index int) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   recordID,
		RecordData: map[string]any{
			"first_name":   gofakeit.FirstName(),
			"last_name":    gofakeit.LastName(),
			"company_name": gofakeit.Company(),
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("error updating contact: %w", err)
	}

	slog.Info("Contact updated successfully",
		"index", index,
		"recordId", res.RecordId,
		"success", res.Success)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreateTag(ctx context.Context, conn *aircall.Connector) (string, error) {
	// Create a unique tag name using gofakeit
	tagName := fmt.Sprintf("%s Tag", gofakeit.BuzzWord())

	recordData := map[string]any{
		"name":  tagName,
		"color": gofakeit.HexColor(),
	}

	slog.Info("Creating tag with data", "data", recordData)

	params := common.WriteParams{
		ObjectName: "tags",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating tag: %w", err)
	}

	slog.Info("Tag created successfully",
		"recordId", res.RecordId,
		"success", res.Success)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdateTag(ctx context.Context, conn *aircall.Connector, recordID string) error {
	params := common.WriteParams{
		ObjectName: "tags",
		RecordId:   recordID,
		RecordData: map[string]any{
			"color": gofakeit.HexColor(),
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("error updating tag: %w", err)
	}

	slog.Info("Tag updated successfully",
		"recordId", res.RecordId,
		"success", res.Success)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	return nil
}
