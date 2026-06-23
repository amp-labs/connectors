package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/servicenow"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	conn := servicenow.GetServiceNowConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"incident", "problem", "contact", "consumer"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
