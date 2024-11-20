package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetOutreachConnector(ctx)

	var err error

	err = testReadSequences(ctx, conn)
	if err != nil {
		utils.Fail("error reading sequences", "error", err)
	}

	err = testReadMailings(ctx, conn)
	if err != nil {
		utils.Fail("error reading mailings", "error", err)
	}

	err = testReadProspects(ctx, conn)
	if err != nil {
		utils.Fail("error reading prospects", "error", err)
	}
}

func testReadSequences(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "sequences",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("openCount", "description", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadMailings(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "mailings",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("bodyHtml", "errorReason", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadProspects(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "prospects",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("addressCountry", "campaignName", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
