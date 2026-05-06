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
	pd "github.com/amp-labs/connectors/providers/procore"
	"github.com/amp-labs/connectors/test/procoresandbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// We are using the Procore sandbox environment because we don't have access to a real Procore instance.
	// The sandbox provides a realistic environment for testing our connector without risking any real data.
	conn, err := procoresandbox.NewConnector(ctx)
	if err != nil {
		slog.Error("Failed to create connector", slog.Any("error", err))
		return
	}

	if _, err := testRead(ctx, conn, "companies", []string{"id", "is_active", "name"}, 1000, ""); err != nil {
		slog.Error(err.Error())
	}

	if _, err := testRead(ctx, conn, "submittal_statuses", []string{"id", "name", "status"}, 4, ""); err != nil {
		slog.Error(err.Error())
	}

	if _, err := testRead(ctx, conn, "vendors", []string{"id", "name", "is_active"}, 1000, ""); err != nil {
		slog.Error(err.Error())
	}

	res, err := testRead(ctx, conn, "checklist/list_templates", []string{"id", "name", "nupdated_atme"}, 2, "")
	if err != nil {
		slog.Error(err.Error())
	}

	if _, err := testRead(ctx, conn, "checklist/list_templates", []string{"id", "name", "updated_at"}, 2, res.NextPage.String()); err != nil {
		slog.Error(err.Error())
	}

}

func testRead(ctx context.Context, conn *pd.Connector, objectName string, fields []string, pageSize int, nextpage string) (*common.ReadResult, error) {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      time.Now().AddDate(0, 0, -30), // 30 days ago
		Until:      time.Now(),
		PageSize:   pageSize,
		NextPage:   common.NextPageToken(nextpage),
	}

	if nextpage == "" {
		log.Printf("Testing read for object %s with fields %v\n", objectName, fields)
	} else {
		log.Printf("Testing read for object %s with fields %v on next page\n", objectName, fields)
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", objectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	if _, err := os.Stdout.Write(jsonStr); err != nil {
		return nil, fmt.Errorf("error writing to stdout: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return nil, fmt.Errorf("error writing to stdout: %w", err)
	}

	log.Printf("Successfully read %d records for object %s\n", len(res.Data), objectName)

	return res, nil
}
