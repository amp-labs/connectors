package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/instantly"
	connTest "github.com/amp-labs/connectors/test/instantly"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

type tagsPayload struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

var objectName = "tags" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetInstantlyConnector(ctx)
	defer utils.Close(conn)

	slog.Info("> TEST Create/Update/Delete tags")
	slog.Info("Creating tags")

	// NOTE: list view must have unique `Name`
	view := createTags(ctx, conn, &tagsPayload{
		Label:       "Tag label",
		Description: "Some description goes here",
	})

	slog.Info("Updating some tags properties")
	updateTags(ctx, conn, view.RecordId, &tagsPayload{
		Label:       "Updated label",
		Description: "Even better description",
	})

	slog.Info("View that tags has changed accordingly")

	res := readTags(ctx, conn)

	updatedView := searchTags(res, "id", view.RecordId)
	for k, v := range map[string]string{
		"label":       "Updated label",
		"description": "Even better description",
	} {
		if !mockutils.DoesObjectCorrespondToString(updatedView[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedView[k])
		}
	}

	slog.Info("Removing this tags")
	removeTags(ctx, conn, view.RecordId)
	slog.Info("> Successful test completion")
}

func searchTags(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding tags")

	return nil
}

func readTags(ctx context.Context, conn *instantly.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "view", "name",
		),
	})
	if err != nil {
		utils.Fail("error reading from Instantly", "error", err)
	}

	return res
}

func createTags(ctx context.Context, conn *instantly.Connector, payload *tagsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Instantly", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a tags")
	}

	return res
}

func updateTags(ctx context.Context, conn *instantly.Connector, viewID string, payload *tagsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Instantly", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a tags")
	}

	return res
}

func removeTags(ctx context.Context, conn *instantly.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Instantly", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a tags")
	}
}
