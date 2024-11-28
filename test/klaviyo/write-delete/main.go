package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/klaviyo"
	connTest "github.com/amp-labs/connectors/test/klaviyo"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/brianvoe/gofakeit/v6"
)

var objectName = "tags"

type TagPayload struct {
	Name string `json:"name"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKlaviyoConnector(ctx)

	slog.Info("> TEST Create/Update/Delete tag")
	slog.Info("Creating tag")

	name := gofakeit.AppName()
	createTag(ctx, conn, &TagPayload{
		Name: name,
	})

	slog.Info("Reading tags")

	res := readTags(ctx, conn)

	slog.Info("Finding recently created tag")

	tag := searchTags(res, "name", name)
	tagID := fmt.Sprintf("%v", tag["id"])

	slog.Info("Updating tag name")

	newName := gofakeit.AppName()
	updateTag(ctx, conn, tagID, &TagPayload{
		Name: newName,
	})

	slog.Info("View that tag has changed accordingly")

	res = readTags(ctx, conn)

	tag = searchTags(res, "id", tagID)
	if tagName, ok := tag["name"].(string); !ok || tagName != newName {
		utils.Fail("error updated name doesn't match")
	}

	slog.Info("Removing this tag")
	removeTag(ctx, conn, tagID)
	slog.Info("> Successful test completion")
}

func searchTags(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding tag")

	return nil
}

func readTags(ctx context.Context, conn *klaviyo.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "name",
		),
	})
	if err != nil {
		utils.Fail("error reading from Klaviyo", "error", err)
	}

	return res
}

func createTag(ctx context.Context, conn *klaviyo.Connector, payload *TagPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Klaviyo", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a tag")
	}
}

func updateTag(ctx context.Context, conn *klaviyo.Connector, tagID string, payload *TagPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   tagID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Klaviyo", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a tag")
	}
}

func removeTag(ctx context.Context, conn *klaviyo.Connector, tagID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   tagID,
	})
	if err != nil {
		utils.Fail("error deleting for Klaviyo", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a tag")
	}
}
