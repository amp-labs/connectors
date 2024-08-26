package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/smartlead"
	connTest "github.com/amp-labs/connectors/test/smartlead"
	"github.com/amp-labs/connectors/test/utils"
)

type campaignPayload struct {
	Name string `json:"name,omitempty"`
}

var (
	objectName = "campaigns" // nolint: gochecknoglobals
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSmartleadConnector(ctx)
	defer utils.Close(conn)

	fmt.Println("> TEST Create/Delete campaign")
	fmt.Println("Creating campaign")

	view := createCampaign(ctx, conn, &campaignPayload{
		Name: "Most recent email campaign",
	})
	utils.DumpJSON(view, os.Stdout)

	fmt.Println("Removing this campaign")
	removeCampaign(ctx, conn, view.RecordId)
	fmt.Println("> Successful test completion")
}

func createCampaign(ctx context.Context, conn *smartlead.Connector, payload *campaignPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Smartlead", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a campaign")
	}

	return res
}

func removeCampaign(ctx context.Context, conn *smartlead.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Smartlead", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a campaign")
	}
}
