package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/solarwinds"
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

	conn := solarwinds.GetSolarWindsConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"incidents", "problems", "departments"})
	if err != nil {
		return err
	}

	// Print the results
	utils.DumpJSON(m.Result, os.Stdout)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
