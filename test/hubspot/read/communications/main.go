package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetHubspotConnector(ctx)

	var since time.Time
	//since = time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC)

	// Object "custom-channels" is excluded because,
	// it requires "Sales Hub Professional" license

	params := []common.ReadParams{
		{
			ObjectName: "communication-preferences",
			Fields:     connectors.Fields("id"),
			Since:      since,
		},
		{
			ObjectName: "channel-accounts",
			Fields:     connectors.Fields("id"),
			Since:      since,
		},
		{
			ObjectName: "channels",
			Fields:     connectors.Fields("id"),
			Since:      since,
		},
		{
			ObjectName: "inboxes",
			Fields:     connectors.Fields("id"),
			Since:      since,
		},
		{
			ObjectName: "threads",
			Fields:     connectors.Fields("id"),
			Since:      since,
		},
	}

	for _, param := range params {
		read(conn, ctx, param)
		fmt.Println("=========================")
	}
}

func read(conn *hubspot.Connector, ctx context.Context, params common.ReadParams) {
	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
