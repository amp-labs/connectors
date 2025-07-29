package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	br "github.com/amp-labs/connectors/providers/braze"
	"github.com/amp-labs/connectors/test/braze"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := braze.NewBrazeConnector(ctx)

	if err := testReadEvents(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadCampaigns(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadEmailTemplates(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadCampaigns(ctx context.Context, conn *br.Connector) error {

	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("id", "name"),
		Since:      time.Now().Add(-13000 * time.Hour),
		// NextPage:   "https://rest.iad-03.braze.com/campaigns/list?page=1",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadEvents(ctx context.Context, conn *br.Connector) error {
	params := common.ReadParams{
		ObjectName: "events",
		Fields:     connectors.Fields("description", "name"),
		// NextPage:   "https://rest.iad-03.braze.com/events/?cursor=c2tpcDo1MA==",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadEmailTemplates(ctx context.Context, conn *br.Connector) error {
	params := common.ReadParams{
		ObjectName: "templates/email",
		Fields:     connectors.Fields("template_name", "email_template_id"),
		Since:      time.Now().Add(-13000 * time.Hour),
		NextPage:   "https://rest.iad-03.braze.com/templates/email/list?offset=101",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
