package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/breezy"
	connTest "github.com/amp-labs/connectors/test/breezy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBreezyConnector(ctx)

	title := fmt.Sprintf("Amp integration %s", gofakeit.Word())

	createPayload := map[string]any{
		"name":        title,
		"type":        "fullTime",
		"description": "Temporary connector integration test position.",
		"location": map[string]any{
			"country":   "US",
			"state":     "CA",
			"city":      "San Francisco",
			"is_remote": true,
		},
		"department":  "Engineering",
		"category":    "software",
		"experience":  "mid-level",
		"pipeline_id": "default",
	}

	updatePayload := map[string]any{
		"name":        title + " (Updated)",
		"type":        "fullTime",
		"description": "Updated via connector integration test.",
		"location": map[string]any{
			"country":   "US",
			"state":     "CA",
			"city":      "San Francisco",
			"is_remote": true,
		},
		"department":  "Engineering",
		"category":    "software",
		"experience":  "senior-level",
		"pipeline_id": "default",
	}

	slog.Info("=== positions (create -> publish -> read -> update -> read -> delete) ===")

	positionID, err := createPosition(ctx, conn, createPayload)
	if err != nil {
		utils.Fail("create failed", "error", err)
	}
	defer func() {
		if positionID == "" {
			return
		}
		if err := archivePosition(ctx, conn, positionID); err != nil {
			slog.Warn("cleanup archive failed", "position_id", positionID, "error", err)
		}
	}()

	slog.Info("=== publish draft position (required for default positions list) ===")
	if err := connTest.PublishPosition(ctx, conn, positionID); err != nil {
		utils.Fail("publish failed", "error", err, "position_id", positionID)
	}
	slog.Info("Position published", "position_id", positionID)

	slog.Info("=== read positions after create ===")
	readRes, err := readPositions(ctx, conn)
	if err != nil {
		utils.Fail("read after create failed", "error", err)
	}
	utils.DumpJSON(readRes, os.Stdout)

	slog.Info("=== update position ===")
	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "positions",
		RecordId:   positionID,
		RecordData: updatePayload,
	})
	if err != nil {
		utils.Fail("update failed", "error", err, "position_id", positionID)
	}
	slog.Info("Position updated", "position_id", positionID)
	utils.DumpJSON(updateRes, os.Stdout)

	slog.Info("=== read positions after update ===")
	readRes, err = readPositions(ctx, conn)
	if err != nil {
		utils.Fail("read after update failed", "error", err)
	}
	utils.DumpJSON(readRes, os.Stdout)

	record, err := findPosition(readRes, positionID)
	if err != nil {
		utils.Fail("position not found after update", "error", err, "position_id", positionID)
	}
	slog.Info("Matched position after update", "position_id", positionID)
	utils.DumpJSON(record, os.Stdout)

	expectedName := title + " (Updated)"
	if !mockutils.DoesObjectCorrespondToString(record.Fields["name"], expectedName) {
		utils.Fail("updated name mismatch", "expected", expectedName, "actual", record.Fields["name"])
	}
	slog.Info("Update validated", "name", record.Fields["name"])

	slog.Info("=== archive position (connector delete) ===")
	deleteRes, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "positions",
		RecordId:   positionID,
	})
	if err != nil {
		utils.Fail("delete failed", "error", err, "position_id", positionID)
	}
	slog.Info("Position archived", "position_id", positionID)
	utils.DumpJSON(deleteRes, os.Stdout)

	positionID = ""

	slog.Info("Breezy positions write-delete test completed successfully")
}

func createPosition(ctx context.Context, conn *breezy.Connector, payload map[string]any) (string, error) {
	slog.Info("Creating position", "payload", payload)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "positions",
		RecordData: payload,
	})
	if err != nil {
		return "", err
	}

	slog.Info("Position created", "record_id", res.RecordId, "success", res.Success)
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("create returned empty RecordId")
	}

	return res.RecordId, nil
}

func readPositions(ctx context.Context, conn *breezy.Connector) (*common.ReadResult, error) {
	fields := datautils.NewSet("_id", "name", "state")

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "positions",
		Fields:     fields,
	})
	if err != nil {
		return nil, err
	}

	for len(res.NextPage) > 0 {
		next, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "positions",
			Fields:     fields,
			NextPage:   res.NextPage,
		})
		if err != nil {
			return nil, err
		}

		res.Data = append(res.Data, next.Data...)
		res.NextPage = next.NextPage
	}

	slog.Info("Read positions", "rows", res.Rows)

	return res, nil
}

func findPosition(res *common.ReadResult, positionID string) (*common.ReadResultRow, error) {
	for _, row := range res.Data {
		if row.Id == positionID {
			return &row, nil
		}

		if mockutils.DoesObjectCorrespondToString(row.Fields["_id"], positionID) {
			return &row, nil
		}
	}

	return nil, fmt.Errorf("position %q not in read list", positionID)
}

func archivePosition(ctx context.Context, conn *breezy.Connector, positionID string) error {
	slog.Info("Archiving position (cleanup)", "position_id", positionID)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "positions",
		RecordId:   positionID,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
