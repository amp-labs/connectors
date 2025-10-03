package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers/outreach"
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

	err := testReadProspects(ctx, conn)
	if err != nil {
		utils.Fail("error reading prospects", "error", err)
	}

	err = testReadMailings(ctx, conn)
	if err != nil {
		utils.Fail("error reading prospects", "error", err)
	}
}

func testReadProspects(ctx context.Context, conn *outreach.Connector) error {
	config := connectors.ReadParams{
		ObjectName:        "prospects",
		Fields:            connectors.Fields("addressCountry", "campaignName", "id"),
		AssociatedObjects: []string{"phoneNumbers", "account", "creator"},
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

func testReadMailings(ctx context.Context, conn *outreach.Connector) error {
	config := connectors.ReadParams{
		ObjectName:        "mailings",
		Since:             time.Now().Add(-720 * time.Hour),
		Fields:            connectors.Fields("bodyHtml", "errorReason", "id"),
		AssociatedObjects: []string{"mailbox", "user", "sequence"},
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
