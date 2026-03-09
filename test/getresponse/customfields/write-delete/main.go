package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/getresponse"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

// MainFn runs the custom-fields write-delete E2E: create custom field → update → delete.
func MainFn() int {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Test 1: Create custom field.
	slog.Info("=== Test 1: Create custom field ===")
	customFieldID, err := createCustomField(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create custom field", "error", err)
	}
	slog.Info("Custom field created", "customFieldId", customFieldID)

	time.Sleep(2 * time.Second)

	// Test 2: Update custom field.
	slog.Info("=== Test 2: Update custom field ===")
	if err := updateCustomField(ctx, conn, customFieldID); err != nil {
		utils.Fail("Failed to update custom field", "error", err)
	}
	slog.Info("Custom field updated")

	// Test 3: Delete custom field (cleanup).
	slog.Info("=== Test 3: Delete custom field ===")
	if err := deleteCustomField(ctx, conn, customFieldID); err != nil {
		utils.Fail("Failed to delete custom field", "error", err)
	}
	slog.Info("Custom field deleted")

	slog.Info("Custom-fields write-delete tests completed successfully!")
	return 0
}

func createCustomField(ctx context.Context, conn *getresponse.Connector) (string, error) {
	name := fmt.Sprintf("TestField_%s", gofakeit.Word())
	recordData := map[string]any{
		"name":      name,
		"valueType": "string",
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: recordData,
	})
	if err != nil {
		return "", err
	}

	utils.DumpJSON(res, os.Stdout)

	if res.RecordId != "" {
		return res.RecordId, nil
	}

	time.Sleep(2 * time.Second)
	return findCustomFieldByName(ctx, conn, name)
}

func findCustomFieldByName(ctx context.Context, conn *getresponse.Connector, name string) (string, error) {
	params := common.ReadParams{
		ObjectName: "custom-fields",
		Fields:     connectors.Fields("customFieldId", "name", "valueType"),
		PageSize:   100,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", err
	}

	for _, row := range res.Data {
		if n, ok := row.Raw["name"].(string); ok && n == name {
			if id, ok := row.Raw["customFieldId"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("custom field not found after create")
}

func updateCustomField(ctx context.Context, conn *getresponse.Connector, customFieldID string) error {
	recordData := map[string]any{
		"name": fmt.Sprintf("Updated_%s", gofakeit.Word()),
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "custom-fields",
		RecordId:   customFieldID,
		RecordData: recordData,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if res.RecordId != customFieldID {
		return fmt.Errorf("expected customFieldId %s, got %s", customFieldID, res.RecordId)
	}

	return nil
}

func deleteCustomField(ctx context.Context, conn *getresponse.Connector, customFieldID string) error {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "custom-fields",
		RecordId:   customFieldID,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if !res.Success {
		return fmt.Errorf("delete reported failure")
	}

	return nil
}
