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
		"contacts/blacklist",
		"texts",
		"phone-numbers",
		"webhooks",
		"tags",
		"messages",
		"user_groups",
		"agent",
		"campaigns",
		"custom-fields",
		"list",
		"number",
		"templates",
		"threads",
		"calls_ai",
		"meetings_ai",
		"sales_dialer/calls",
		"sales_dialer/contacts",
		"sales_dialer/campaigns/contacts",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("=== JustCall Metadata ===")
	utils.DumpJSON(res, os.Stdout)
}
