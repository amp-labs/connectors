package testscenario

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
)

// ReadThroughPages reads records from a connector page by page.
// It continues advancing pagination until no more pages remain.
// This is useful in tests to ensure pagination terminates and does not loop infinitely.
func ReadThroughPages(ctx context.Context, connector connectors.ReadConnector, params common.ReadParams) {
	var (
		read = &common.ReadResult{
			Done: false, // seed initial state so the loop runs at least once
		}
		err error
	)

	numPages := 0

	for !read.Done {
		// Advance pagination.
		params.NextPage = read.NextPage

		read, err = connector.Read(ctx, params)
		if err != nil {
			utils.Fail("error reading from connector", "error", err)
		}

		// Log and dump results for inspection.
		slog.Info("Reading...", "page", params.NextPage)
		utils.DumpJSON(read, os.Stdout)

		numPages += 1
	}

	slog.Info("Number of pages read", "numPages", numPages)
}

// SearchThroughPages searches records from a connector page by page.
// It continues advancing pagination until no more pages remain.
// This is useful in tests to ensure pagination terminates and does not loop infinitely.
func SearchThroughPages(ctx context.Context, connector connectors.SearchConnector, params common.SearchParams) {
	var (
		read = &common.SearchResult{
			Done: false, // seed initial state so the loop runs at least once
		}
		err error
	)

	numPages := 0

	for !read.Done {
		// Advance pagination.
		params.NextPage = read.NextPage

		read, err = connector.Search(ctx, &params)
		if err != nil {
			utils.Fail("error reading from connector", "error", err)
		}

		// Log and dump results for inspection.
		slog.Info("Reading...", "page", params.NextPage)
		utils.DumpJSON(read, os.Stdout)

		numPages += 1
	}

	slog.Info("Number of pages read", "numPages", numPages)
}
