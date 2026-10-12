package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCalendlyConnector(ctx)

	fmt.Println("Resolving post-auth info...")

	if _, err := conn.GetPostAuthInfo(ctx); err != nil {
		utils.Fail("GetPostAuthInfo", "error", err)
	}

	fmt.Println("Reading event_types page...")

	readRes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "event_types",
		Fields:     connectors.Fields("uri", "name"),
	})
	if err != nil {
		utils.Fail("Read event_types", "error", err)
	}

	if len(readRes.Data) == 0 {
		utils.Fail("no event types returned; cannot exercise GetRecordsByIds")
	}

	uri, _ := readRes.Data[0].Fields["uri"].(string)
	if uri == "" {
		if r, ok := readRes.Data[0].Raw["uri"].(string); ok {
			uri = r
		}
	}

	if uri == "" {
		utils.Fail("could not resolve event type URI from first row")
	}

	fmt.Println("\nFetching event type by URI via GetRecordsByIds...")

	res, err := conn.GetRecordsByIds(ctx, "event_types", []string{uri},
		[]string{"uri", "name", "active"}, nil)
	if err != nil {
		utils.Fail("GetRecordsByIds", "error", err)
	}

	fmt.Printf("\nSuccessfully fetched %d event type row(s):\n", len(res))
	utils.DumpJSON(res, os.Stdout)

	if len(res) != 1 {
		utils.Fail("expected exactly one row from GetRecordsByIds", "got", len(res))
	}

	gotURI, _ := res[0].Fields["uri"].(string)
	if gotURI == "" {
		if r, ok := res[0].Raw["uri"].(string); ok {
			gotURI = r
		}
	}

	if gotURI != uri {
		utils.Fail("URI mismatch", "expected", uri, "got", gotURI)
	}

	fmt.Println("\n✓ GetRecordsByIds returned the expected event type.")
}
