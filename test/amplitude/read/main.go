package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	amplitudetest "github.com/amp-labs/connectors/test/amplitude"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := amplitudetest.GetAmplitudeConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "events",
		Fields:     connectors.Fields("id", "autohidden", "display"),
	})
	if err != nil {
		utils.Fail("error reading from amplitude", "error", err)
	}

	slog.Info("Reading events..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "annotations",
		Fields:     connectors.Fields("id", "data", "label"),
	})
	if err != nil {
		utils.Fail("error reading from amplitude", "error", err)
	}

	slog.Info("Reading annotations..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "taxonomy/category",
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from amplitude", "error", err)
	}

	slog.Info("Reading taxonomy/category..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "taxonomy/user-property",
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from amplitude", "error", err)
	}

	slog.Info("Reading taxonomy/user-property..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
