package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	sm "github.com/amp-labs/connectors/providers/seismic"
	"github.com/amp-labs/connectors/test/seismic"
	"github.com/amp-labs/connectors/test/utils"
)

var readSupport = []string{
	// These are supported objects in the reporting API,
	// ref: https://developer.seismic.com/seismicsoftware/reference/h1-reporting-api-overview
	"adminImpersonationSessions",
	"aiActivity",
	"aiGeneratedText",
	"aiGeneratedTextUserFeedback",
	"aiSuggestedContentProperties",
	"announcements",
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := seismic.GetConnector(ctx)
	for _, objectName := range readSupport {
		if err := testRead(ctx, conn, objectName, []string{"id"}); err != nil {
			slog.Error(err.Error())
		}
	}
}

func testRead(ctx context.Context, conn *sm.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	if _, err := os.Stdout.Write(jsonStr); err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	return nil
}
