package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/sageintacct"
	connTest "github.com/amp-labs/connectors/test/sageintacct"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "account"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSageIntacctConnector(ctx)

	slog.Info("> TEST Create/Delete accounts")
	slog.Info("Creating account")

	accountID, err := createSegment(ctx, conn)
	if err != nil {
		slog.Error("Failed to create account", "error", err)
		return
	}

	slog.Info("Removing this account")
	deleteAccount(ctx, conn, accountID)
	slog.Info("> Successful test completion")
}

func createSegment(ctx context.Context, conn *sageintacct.Connector) (string, error) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: map[string]any{
			"id":                    "15277",
			"name":                  "Vehicle parts - Transmission",
			"accountType":           "balanceSheet",
			"closingType":           "nonClosingAccount",
			"normalBalance":         "debit",
			"alternativeGLAccount":  "none",
			"status":                "active",
			"isTaxable":             false,
			"disallowDirectPosting": true,
		},
	})
	if err != nil {
		utils.Fail("error writing to Customer App", "error", err)
		return "", err
	}

	if !res.Success {
		utils.Fail("failed to create a segment")
		return "", fmt.Errorf("failed to create a segment")
	}

	return res.RecordId, nil
}

func deleteAccount(ctx context.Context, conn *sageintacct.Connector, segmentID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   segmentID,
	})
	if err != nil {
		utils.Fail("error deleting for Sage Intacct", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a segment")
	}
}
