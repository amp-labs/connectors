package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/supersend"
	testsupersend "github.com/amp-labs/connectors/test/supersend"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
	ctx := context.Background()
	conn := testsupersend.GetSuperSendConnector(ctx)

	// Step 1: Create a team first (needed for other operations)
	teamID, err := testCreateTeam(ctx, conn)
	if err != nil {
		slog.Error("create team failed", "error", err)
		return 1
	}

	// Step 2: Create a label using the team ID
	labelID, err := testCreateLabel(ctx, conn, teamID)
	if err != nil {
		slog.Error("create label failed", "error", err)
		return 1
	}

	// Step 3: Update the label
	err = testUpdateLabel(ctx, conn, labelID)
	if err != nil {
		slog.Error("update label failed", "error", err)
		return 1
	}

	slog.Info("All write tests completed successfully!")

	return 0
}

func testCreateTeam(ctx context.Context, conn *supersend.Connector) (string, error) {
	slog.Info("Testing create team...")

	params := common.WriteParams{
		ObjectName: "teams",
		RecordData: map[string]any{
			"name":   fmt.Sprintf("Test Team %d", os.Getpid()),
			"domain": "testteam.example.com",
			"about":  "Created via write test",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error("error writing team", "error", err)
		return "", err
	}

	slog.Info("Create team response:")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("create team completed", "success", res.Success, "recordId", res.RecordId)

	return res.RecordId, nil
}

func testCreateLabel(ctx context.Context, conn *supersend.Connector, teamID string) (string, error) {
	slog.Info("Testing create label...", "teamId", teamID)

	params := common.WriteParams{
		ObjectName: "labels",
		RecordData: map[string]any{
			"name":   "Test Label",
			"color":  "#FF5733",
			"TeamId": teamID,
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error("error writing label", "error", err)
		return "", err
	}

	slog.Info("Create label response:")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("create label completed", "success", res.Success, "recordId", res.RecordId)

	return res.RecordId, nil
}

func testUpdateLabel(ctx context.Context, conn *supersend.Connector, labelID string) error {
	slog.Info("Testing update label...", "labelId", labelID)

	params := common.WriteParams{
		ObjectName: "labels",
		RecordId:   labelID,
		RecordData: map[string]any{
			"name":  "Updated Test Label",
			"color": "#00FF00",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error("error updating label", "error", err)
		return err
	}

	slog.Info("Update label response:")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("update label completed", "success", res.Success, "recordId", res.RecordId)

	return nil
}
