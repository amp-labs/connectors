package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	testJustCall "github.com/amp-labs/connectors/test/justcall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testJustCall.GetJustCallConnector(ctx)

	res, err := conn.ListObjectMetadata(ctx, []string{
		"users",
		"calls",
		"contacts",
		"texts",
		"phone-numbers",
		"webhooks",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("=== JustCall Metadata ===")
	utils.DumpJSON(res, os.Stdout)
}
