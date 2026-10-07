package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/freshdesk"
	"github.com/amp-labs/connectors/test/utils"
)

// objects mirrors the read supported objects of the connector.
var objects = []string{ //nolint:gochecknoglobals
	"agents",
	"companies",
	"contacts",
	"products",
	"scenario_automations",
	"settings",
	"ticket-forms",
	"tickets",
}

func main() {
	ctx := context.Background()

	conn := freshdesk.GetFreshdeskConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		slog.Error(err.Error())

		return
	}

	utils.DumpJSON(m, os.Stdout)
}
