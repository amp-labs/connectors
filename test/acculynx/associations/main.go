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
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// Live validation for the representative/estimate associations and estimate
// hydration work:
//
//  1. Read jobs/representatives with AssociatedObjects=[jobs] and show every
//     representative row carries a "jobs" edge with the fan-out parent job id
//     (the representative body itself has no jobId, only _link).
//  2. Read estimates with AssociatedObjects=[jobs] and show every estimate row
//     carries a "jobs" edge taken from the embedded job stub, with isPrimary
//     in ProviderAssociationMetadata.
//  3. Read estimates and show rows are hydrated from GET /estimates/{id}:
//     estimateNumber, createdDate and financials populate Fields even though
//     the list payload is a reference stub (verified: ?includes= is ignored
//     on the list endpoint).
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("=== 1) Read jobs/representatives with the parent job association ===")

	if err := readRepresentativesWithJobEdge(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("=== 2) Read estimates with the job association ===")

	if err := readEstimatesWithJobEdge(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("=== 3) Read estimates hydrated from the detail endpoint ===")

	if err := readEstimatesHydrated(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readRepresentativesWithJobEdge(ctx context.Context, conn *acculynx.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "jobs/representatives",
		Fields:            connectors.Fields("id", "type"),
		AssociatedObjects: []string{"jobs"},
	})
	if err != nil {
		return fmt.Errorf("reading jobs/representatives with associations: %w", err)
	}

	withJob := 0

	for _, row := range res.Data {
		if len(row.Associations["jobs"]) > 0 {
			withJob++
		}
	}

	slog.Info("representative association coverage",
		"rows", res.Rows,
		"withJobEdge", withJob,
		"expected", "withJobEdge == rows")

	dumpFirstRows(res)

	return nil
}

func readEstimatesWithJobEdge(ctx context.Context, conn *acculynx.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "estimates",
		Fields:            connectors.Fields("id"),
		AssociatedObjects: []string{"jobs"},
	})
	if err != nil {
		return fmt.Errorf("reading estimates with associations: %w", err)
	}

	withJob, withIsPrimary := 0, 0

	for _, row := range res.Data {
		edges := row.Associations["jobs"]
		if len(edges) > 0 {
			withJob++

			if _, ok := edges[0].ProviderAssociationMetadata["isPrimary"]; ok {
				withIsPrimary++
			}
		}
	}

	slog.Info("estimate association coverage",
		"rows", res.Rows,
		"withJobEdge", withJob,
		"withIsPrimaryMetadata", withIsPrimary,
		"expected", "both == rows")

	dumpFirstRows(res)

	return nil
}

func readEstimatesHydrated(ctx context.Context, conn *acculynx.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "estimates",
		Fields:     connectors.Fields("id", "estimateNumber", "createdDate", "financials"),
	})
	if err != nil {
		return fmt.Errorf("reading estimates hydrated: %w", err)
	}

	hydrated := 0

	for _, row := range res.Data {
		if row.Fields["estimatenumber"] != nil && row.Fields["financials"] != nil {
			hydrated++
		}
	}

	slog.Info("estimate hydration coverage",
		"rows", res.Rows,
		"hydrated", hydrated,
		"expected", "hydrated == rows")

	dumpFirstRows(res)

	return nil
}

func dumpFirstRows(res *common.ReadResult) {
	limit := 3
	if len(res.Data) < limit {
		limit = len(res.Data)
	}

	utils.DumpJSON(res.Data[:limit], os.Stdout)
}
