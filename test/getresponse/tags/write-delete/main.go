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

// MainFn runs the tags write-delete E2E: create tag → update → delete.
func MainFn() int {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Test 1: Create tag.
	slog.Info("=== Test 1: Create tag ===")
	tagID, err := createTag(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create tag", "error", err)
	}
	slog.Info("Tag created", "tagId", tagID)

	time.Sleep(2 * time.Second)

	// Test 2: Update tag.
	slog.Info("=== Test 2: Update tag ===")
	if err := updateTag(ctx, conn, tagID); err != nil {
		utils.Fail("Failed to update tag", "error", err)
	}
	slog.Info("Tag updated")

	// Test 3: Delete tag (cleanup).
	slog.Info("=== Test 3: Delete tag ===")
	if err := deleteTag(ctx, conn, tagID); err != nil {
		utils.Fail("Failed to delete tag", "error", err)
	}
	slog.Info("Tag deleted")

	slog.Info("Tags write-delete tests completed successfully!")
	return 0
}

// tagName returns a name valid for GetResponse tags: only [A-Za-z0-9_] allowed.
func tagName(prefix string) string {
	return prefix + gofakeit.Numerify("########")
}

func createTag(ctx context.Context, conn *getresponse.Connector) (string, error) {
	name := tagName("TestTag_")
	recordData := map[string]any{
		"name":  name,
		"color": "#3498db",
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
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
	return findTagByName(ctx, conn, name)
}

func findTagByName(ctx context.Context, conn *getresponse.Connector, name string) (string, error) {
	params := common.ReadParams{
		ObjectName: "tags",
		Fields:     connectors.Fields("tagId", "name", "createdAt"),
		Filter:     fmt.Sprintf("query[name]=%s", name),
		PageSize:   10,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", err
	}

	for _, row := range res.Data {
		if n, ok := row.Raw["name"].(string); ok && n == name {
			if id, ok := row.Raw["tagId"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("tag not found after create")
}

func updateTag(ctx context.Context, conn *getresponse.Connector, tagID string) error {
	recordData := map[string]any{
		"name":  tagName("Updated_"),
		"color": "#e74c3c",
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordId:   tagID,
		RecordData: recordData,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if res.RecordId != tagID {
		return fmt.Errorf("expected tagId %s, got %s", tagID, res.RecordId)
	}

	return nil
}

func deleteTag(ctx context.Context, conn *getresponse.Connector, tagID string) error {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "tags",
		RecordId:   tagID,
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
