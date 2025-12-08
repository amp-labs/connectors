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

	// Main API objects
	objects := []string{
		"users",
		"calls",
		"contacts",
		"texts",
		"phone-numbers",
		"webhooks",
		"texts/tags",
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
