package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/amplitude"
	amplitudetest "github.com/amp-labs/connectors/test/amplitude"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := amplitudetest.GetAmplitudeConnector(ctx)

	readAndLog(ctx, conn, amplitude.AnnotationsObject)
	readAndLog(ctx, conn, amplitude.CohortsObject)

	slog.Info("Read operation completed successfully.")
}

func readAndLog(ctx context.Context, conn *amplitude.Connector, objectName string) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
	})
	if err != nil {
		utils.Fail("error reading from Amplitude", "object", objectName, "error", err)
	}

	slog.Info("Reading " + objectName + "..")
	utils.DumpJSON(res, os.Stdout)
}
