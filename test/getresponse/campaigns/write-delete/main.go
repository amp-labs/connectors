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

// MainFn runs the campaigns write-delete E2E: create campaign → update → delete.
func MainFn() int {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Test 1: Create campaign.
	slog.Info("=== Test 1: Create campaign ===")
	campaignID, err := createCampaign(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create campaign", "error", err)
	}
	slog.Info("Campaign created", "campaignId", campaignID)

	time.Sleep(2 * time.Second)

	// Test 2: Update campaign.
	slog.Info("=== Test 2: Update campaign ===")
	if err := updateCampaign(ctx, conn, campaignID); err != nil {
		utils.Fail("Failed to update campaign", "error", err)
	}
	slog.Info("Campaign updated")

	// Test 3: Delete campaign (cleanup).
	slog.Info("=== Test 3: Delete campaign ===")
	if err := deleteCampaign(ctx, conn, campaignID); err != nil {
		utils.Fail("Failed to delete campaign", "error", err)
	}
	slog.Info("Campaign deleted")

	slog.Info("Campaigns write-delete tests completed successfully!")
	return 0
}

func createCampaign(ctx context.Context, conn *getresponse.Connector) (string, error) {
	name := fmt.Sprintf("Test Campaign %s", gofakeit.Word())
	recordData := map[string]any{
		"name":         name,
		"languageCode": "EN",
		"description":  fmt.Sprintf("Test at %s", time.Now().Format(time.RFC3339)),
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "campaigns",
		RecordData: recordData,
	})
	if err != nil {
		return "", err
	}

	utils.DumpJSON(res, os.Stdout)

	if res.RecordId != "" {
		return res.RecordId, nil
	}

	// GetResponse may return 202 Accepted without body; find by name after a short delay.
	time.Sleep(3 * time.Second)
	return findCampaignByName(ctx, conn, name)
}

func findCampaignByName(ctx context.Context, conn *getresponse.Connector, name string) (string, error) {
	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Filter:     fmt.Sprintf("query[name]=%s", name),
		PageSize:   10,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", err
	}

	for _, row := range res.Data {
		if n, ok := row.Raw["name"].(string); ok && n == name {
			if id, ok := row.Raw["campaignId"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("campaign not found after create")
}

func updateCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string) error {
	recordData := map[string]any{
		"name":        fmt.Sprintf("Updated %s", gofakeit.Word()),
		"description": fmt.Sprintf("Updated at %s", time.Now().Format(time.RFC3339)),
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "campaigns",
		RecordId:   campaignID,
		RecordData: recordData,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if res.RecordId != campaignID {
		return fmt.Errorf("expected campaignId %s, got %s", campaignID, res.RecordId)
	}

	return nil
}

func deleteCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string) error {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "campaigns",
		RecordId:   campaignID,
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
