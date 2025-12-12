package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/justcall"
	testJustCall "github.com/amp-labs/connectors/test/justcall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testJustCall.GetJustCallConnector(ctx)

	if err := run(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func run(ctx context.Context, conn *justcall.Connector) error {
	// First create a tag, then delete it
	tagID, err := createTestTag(ctx, conn)
	if err != nil {
		return err
	}

	if err := testDeleteTag(ctx, conn, tagID); err != nil {
		return err
	}

	return nil
}

func createTestTag(ctx context.Context, conn *justcall.Connector) (string, error) {
	slog.Info("Creating a test tag for delete test")

	// JustCall has a 15 character limit for tag names
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"name":       fmt.Sprintf("DelTag%d", os.Getpid()%10000),
			"color_code": "#FF0000",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create tag: %w", err)
	}

	printResult("tags (CREATE for delete test)", res)

	if res.RecordId == "" {
		return "", fmt.Errorf("no record ID returned from create")
	}

	return res.RecordId, nil
}

func testDeleteTag(ctx context.Context, conn *justcall.Connector, tagID string) error {
	slog.Info("Deleting the tag", "tagID", tagID)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "tags",
		RecordId:   tagID,
	})
	if err != nil {
		return fmt.Errorf("tags delete: %w", err)
	}

	printDeleteResult("tags (DELETE)", res)

	return nil
}

func printResult(name string, res *common.WriteResult) {
	jsonStr, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("\n=== %s ===\n", name)
	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")
}

func printDeleteResult(name string, res *common.DeleteResult) {
	jsonStr, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("\n=== %s ===\n", name)
	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")
}
