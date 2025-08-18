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
	cl "github.com/amp-labs/connectors/providers/calendly"
	"github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := calendly.GetCalendlyConnector(ctx)

	conn.GetPostAuthInfo(ctx)

	if err := testReadEvents(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadEventTypes(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadGroupRelationships(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadUserBusyTimes(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadEvents(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "scheduled_events",
		Fields:     connectors.Fields("meeting_notes_plain", "name", "uri"),
		Since:      time.Now().Add(-13000 * time.Hour),
		// NextPage:   "https://api.calendly.com/scheduled_events?count=4\u0026organization=https%3A%2F%2Fapi.calendly.com%2Forganizations%2F098ccc5a-9617-41b2-9986-c6691422281c\u0026page_token=8OkBAMZQMV43AmFssK6PqNFpm0eZcpnr\u0026user=https%3A%2F%2Fapi.calendly.com%2Fusers%2F42687819-a60c-446a-b42f-0d84ce589f0e",
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

func testReadEventTypes(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "event_types",
		Fields:     connectors.Fields("uri", "name", "booking_method"),
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

func testReadGroupRelationships(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "group_relationships",
		Fields:     connectors.Fields("uri", "owner"),
		Since:      time.Now().Add(-13000 * time.Hour),
		NextPage:   "",
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

func testReadUserBusyTimes(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "user_busy_times",
		Fields:     connectors.Fields("type", "start_time"),
		Since:      time.Now().Add(1 * time.Hour),
		Until:      time.Now().Add(48 * time.Hour),
		NextPage:   "",
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
