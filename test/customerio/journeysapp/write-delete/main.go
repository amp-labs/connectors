package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/customerapp"
	connTest "github.com/amp-labs/connectors/test/customerio/journeysapp"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "segments"

type SegmentCreatePayload struct {
	Segment Segment `json:"segment"`
}

type Segment struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCustomerJourneysAppConnector(ctx)

	slog.Info("> TEST Create/Delete segments")
	slog.Info("Creating segment")

	description := "created from test, will be removed"
	createSegment(ctx, conn, &SegmentCreatePayload{
		Segment: Segment{
			Name:        "finished",
			Description: description,
		},
	})

	slog.Info("Reading segments")

	res := readSegments(ctx, conn)

	slog.Info("Finding recently created segment")

	segment := searchSegments(res, "description", description)
	segmentID := fmt.Sprintf("%v", segment["id"])

	slog.Info("Removing this segment")
	removeSegment(ctx, conn, segmentID)
	slog.Info("> Successful test completion")
}

func searchSegments(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding segment")

	return nil
}

func readSegments(ctx context.Context, conn *customerapp.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "description",
		),
	})
	if err != nil {
		utils.Fail("error reading from Customer App", "error", err)
	}

	return res
}

func createSegment(ctx context.Context, conn *customerapp.Connector, payload *SegmentCreatePayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Customer App", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a segment")
	}
}

func removeSegment(ctx context.Context, conn *customerapp.Connector, segmentID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   segmentID,
	})
	if err != nil {
		utils.Fail("error deleting for Customer App", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a segment")
	}
}
