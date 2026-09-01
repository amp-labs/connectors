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
	housecallpro "github.com/amp-labs/connectors/providers/housecallPro"
	testHousecallPro "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
)

// Live validation for the embedded customer association:
//
//  1. Read jobs with AssociatedObjects=[customers] and show every row carries a
//     "customers" edge lifted off the embedded customer object (pre-existing
//     behaviour, checked as a regression).
//  2. Read estimates with AssociatedObjects=[customers] and show the same edge.
//  3. Read leads with AssociatedObjects=[customers] and show the same edge.
//  4. Read estimates and leads WITHOUT requesting the association and show no
//     edges are attached.
//
// No extra API calls are involved: the customer is already embedded in each
// list payload, so the edge is lifted during marshalling.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testHousecallPro.GetConnector(ctx)

	for _, tc := range []struct {
		object string
		fields []string
	}{
		{"jobs", []string{"id", "work_status", "updated_at"}},
		{"estimates", []string{"id", "estimate_number", "updated_at"}},
		{"leads", []string{"id", "number", "status"}},
	} {
		slog.Info(fmt.Sprintf("=== %v with customers association requested ===", tc.object))

		if err := readWithCustomerEdge(ctx, conn, tc.object, tc.fields); err != nil {
			slog.Error(err.Error())
		}
	}

	for _, tc := range []struct {
		object string
		fields []string
	}{
		{"estimates", []string{"id", "estimate_number", "updated_at"}},
		{"leads", []string{"id", "number", "status"}},
	} {
		slog.Info(fmt.Sprintf("=== %v without the association requested ===", tc.object))

		if err := readWithoutCustomerEdge(ctx, conn, tc.object, tc.fields); err != nil {
			slog.Error(err.Error())
		}
	}
}

func readWithCustomerEdge(
	ctx context.Context, conn *housecallpro.Connector, objectName string, fields []string,
) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        objectName,
		Fields:            connectors.Fields(fields...),
		PageSize:          5,
		AssociatedObjects: []string{"customers"},
	})
	if err != nil {
		return fmt.Errorf("reading %v with associations: %w", objectName, err)
	}

	withCustomer, withEmbedded, withRawPayload := 0, 0, 0

	for _, row := range res.Data {
		edges := row.Associations["customers"]
		if len(edges) > 0 {
			withCustomer++

			if edges[0].ObjectId != "" {
				withEmbedded++
			}

			if len(edges[0].Raw) > 0 {
				withRawPayload++
			}
		}
	}

	slog.Info("customer association coverage",
		"object", objectName,
		"rows", res.Rows,
		"withCustomerEdge", withCustomer,
		"withObjectId", withEmbedded,
		"withRawPayload", withRawPayload,
		"expected", "all three == rows")

	dumpFirstRows(res)

	return nil
}

func readWithoutCustomerEdge(
	ctx context.Context, conn *housecallpro.Connector, objectName string, fields []string,
) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		PageSize:   5,
	})
	if err != nil {
		return fmt.Errorf("reading %v without associations: %w", objectName, err)
	}

	withAny := 0

	for _, row := range res.Data {
		if len(row.Associations) > 0 {
			withAny++
		}
	}

	slog.Info("association leakage check",
		"object", objectName,
		"rows", res.Rows,
		"rowsWithAnyAssociation", withAny,
		"expected", "rowsWithAnyAssociation == 0")

	return nil
}

func dumpFirstRows(res *common.ReadResult) {
	utils.DumpJSON(res.Data[:min(2, len(res.Data))], os.Stdout)
}
