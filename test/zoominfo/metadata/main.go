package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/providers/zoominfo"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoominfo"
)

// This harness lists metadata for every object the ZoomInfo connector supports.
// Metadata is derived by sampling live data, so objects that require search
// criteria (e.g. contacts, intent, all enrich-*) or are entitlement-gated
// (e.g. GTM config, audiences, agent-teams) will appear under "Errors" rather
// than "Result" — that is expected without the relevant inputs/entitlements.
func main() {
	ctx := context.Background()

	conn := connTest.GetZoomInfoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, zoominfo.SupportedObjectNames())
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
