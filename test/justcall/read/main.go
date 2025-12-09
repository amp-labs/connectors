package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/justcall"
	testJustCall "github.com/amp-labs/connectors/test/justcall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testJustCall.GetJustCallConnector(ctx)

	// All verified working API objects
	objects := []string{
		// Core JustCall objects
		"users",
		"calls",
		"contacts",
		"texts",
		"phone-numbers",
		"webhooks",
		"texts/tags",
		"blacklisted-contacts",
		"whatsapp/messages",
		"user_groups",
		// JustCall AI
		"calls_ai",
		// Sales Dialer (requires Sales Dialer subscription)
		"sales_dialer/calls",
		// Note: sales_dialer/campaigns and sales_dialer/contacts
		// return errors without active Sales Dialer subscription
	}

	for _, obj := range objects {
		if err := testRead(ctx, conn, obj); err != nil {
			slog.Error(err.Error())
		}
	}
}

func testRead(ctx context.Context, conn *justcall.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	fmt.Printf("\n=== %s ===\n", objectName)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
