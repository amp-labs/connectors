package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/jump"
	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// liveTestObjects covers every supported read object.
//
//nolint:gochecknoglobals
var liveTestObjects = []string{
	"contacts",
	"documents",
	"integrations",
	"meetingPreps",
	"meetings",
	"notes",
	"pulses",
	"scorecards",
	"signalDefinitions",
	"tasks",
	"users",
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetJumpConnector(ctx)

	// Every page is printed with the requested Fields and the untouched Raw record.
	for _, params := range readExamples() {
		slog.Info("=== Reading " + params.ObjectName + " ===")
		testscenario.ReadThroughPages(ctx, conn, params)
	}

	// Incremental read: only tasks updated inside the window are returned.
	slog.Info("=== Reading tasks updated between Since and Until ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "tasks",
		Fields:     connectors.Fields("id", "title", "status", "updatedAt"),
		Since:      time.Date(2026, time.June, 23, 23, 18, 9, 145000000, time.UTC),
		Until:      time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC),
	})

	// Contacts paginate with the insertedAfter window instead of cursors, see NextPage.
	slog.Info("=== Reading contacts page by page ===")
	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "name", "email", "status", "insertedAt", "updatedAt"),
		PageSize:   100,
	})

	runChecks(ctx, conn)
}

func readExamples() []common.ReadParams {
	return []common.ReadParams{
		{ObjectName: "documents", Fields: connectors.Fields("id", "name", "processingState", "updatedAt")},
		{ObjectName: "integrations", Fields: connectors.Fields("id", "service", "kind", "accountEmail", "updatedAt")},
		{ObjectName: "meetingPreps", Fields: connectors.Fields("id", "contactId", "userId", "updatedAt")},
		{
			ObjectName: "meetings",
			Fields: connectors.Fields("id", "topic", "status", "source", "startedAt", "duration", "updatedAt",
				"participants"),
		},
		{ObjectName: "notes", Fields: connectors.Fields("id", "subject", "meetingId", "updatedAt")},
		{ObjectName: "pulses", Fields: connectors.Fields("id", "name", "description", "updatedAt")},
		{ObjectName: "scorecards", Fields: connectors.Fields("id", "name", "format", "updatedAt", "criteria")},
		{ObjectName: "signalDefinitions", Fields: connectors.Fields("id", "key", "name", "updatedAt")},
		{ObjectName: "tasks", Fields: connectors.Fields("id", "title", "status", "due", "updatedAt", "assignee")},
		{ObjectName: "users", Fields: connectors.Fields("id", "fullName", "email", "owner", "updatedAt")},
	}
}

// report logs a failed check. A rate limit stops the run: Jump extends the block while requests keep arriving.
func report(err error) {
	if errors.Is(err, common.ErrLimitExceeded) {
		utils.Fail("rate limited by Jump, stopping", "error", err)
	}

	slog.Error(err.Error())
}

// runChecks fails loudly on missed or duplicated records, complexity errors and misfiltered cutoffs.
func runChecks(ctx context.Context, conn *jump.Connector) {
	slog.Info("=== Checks ===")

	// Every field, including the optional nested relations, is requested so the
	// page size chosen by the connector is checked against Jump's complexity limit.
	for _, obj := range liveTestObjects {
		if err := testReadAllFields(ctx, conn, obj); err != nil {
			report(err)
		}
	}

	// Cursor pagination, one record per page.
	if err := testPagination(ctx, conn, "tasks", 1); err != nil {
		report(err)
	}

	// The insertedAfter window for contacts. Ten per page keeps the run well within the rate limit.
	if err := testPagination(ctx, conn, "contacts", 10); err != nil {
		report(err)
	}

	// Provider-side updatedAfter/updatedBefore filtering.
	for _, obj := range []string{"tasks", "signalDefinitions", "contacts"} {
		if err := testIncrementalRead(ctx, conn, obj); err != nil {
			report(err)
		}
	}
}

func testReadAllFields(ctx context.Context, conn *jump.Connector, objectName string) error {
	metadata, err := conn.ListObjectMetadata(ctx, []string{objectName})
	if err != nil {
		return fmt.Errorf("error listing metadata for %s: %w", objectName, err)
	}

	fields := datautils.NewStringSet()
	for field := range metadata.Result[objectName].Fields {
		fields.AddOne(field)
	}

	rows, err := readAll(ctx, conn, common.ReadParams{
		ObjectName: objectName,
		Fields:     fields,
	})
	if err != nil {
		return err
	}

	slog.Info("Read result", "object", objectName, "fields", len(fields), "rows", len(rows))

	return nil
}

func testPagination(ctx context.Context, conn *jump.Connector, objectName string, pageSize int) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
		PageSize:   pageSize,
	}

	seen := datautils.NewStringSet()
	pages := 0

	for {
		res, err := conn.Read(ctx, params)
		if err != nil {
			return fmt.Errorf("error reading %s page %d: %w", objectName, pages+1, err)
		}

		pages++

		for _, row := range res.Data {
			if seen.Has(row.Id) {
				return fmt.Errorf("%s page %d returned duplicate record %s", objectName, pages, row.Id)
			}

			seen.AddOne(row.Id)
		}

		if res.Done {
			break
		}

		params.NextPage = res.NextPage
	}

	slog.Info("Pagination result", "object", objectName, "pages", pages, "records", len(seen))

	return nil
}

// testIncrementalRead checks updatedAt cutoffs taken from the records themselves, so it fails if the connector
// loses precision: Jump compares timestamps to the microsecond and both bounds are inclusive.
func testIncrementalRead(ctx context.Context, conn *jump.Connector, objectName string) error {
	all, err := readAll(ctx, conn, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "updatedAt"),
	})
	if err != nil {
		return err
	}

	updated := make(map[string]time.Time, len(all))
	cutoffs := make([]time.Time, 0, len(all))

	for _, row := range all {
		updatedAt, err := time.Parse(time.RFC3339Nano, fmt.Sprint(row.Fields["updatedat"]))
		if err != nil {
			return fmt.Errorf("%s record %s has no updatedAt: %w", objectName, row.Id, err)
		}

		updated[row.Id] = updatedAt
		cutoffs = append(cutoffs, updatedAt)
	}

	for _, cutoff := range sampleCutoffs(cutoffs) {
		since, err := readIDs(ctx, conn, common.ReadParams{ObjectName: objectName, Since: cutoff})
		if err != nil {
			return err
		}

		until, err := readIDs(ctx, conn, common.ReadParams{ObjectName: objectName, Until: cutoff})
		if err != nil {
			return err
		}

		for id, updatedAt := range updated {
			if since.Has(id) != !updatedAt.Before(cutoff) || until.Has(id) != !updatedAt.After(cutoff) {
				return fmt.Errorf("%s record %s (updated %s) misfiltered at cutoff %s",
					objectName, id, updatedAt.Format(time.RFC3339Nano), cutoff.Format(time.RFC3339Nano))
			}
		}
	}

	slog.Info("Incremental read result", "object", objectName, "records", len(updated))

	return nil
}

// sampleCutoffs keeps the run short for large objects: the earliest and latest cutoffs plus evenly spaced ones.
func sampleCutoffs(cutoffs []time.Time) []time.Time {
	const maxCutoffs = 10

	sort.Slice(cutoffs, func(i, j int) bool { return cutoffs[i].Before(cutoffs[j]) })

	if len(cutoffs) <= maxCutoffs {
		return cutoffs
	}

	sampled := make([]time.Time, 0, maxCutoffs)
	for i := range maxCutoffs {
		sampled = append(sampled, cutoffs[i*(len(cutoffs)-1)/(maxCutoffs-1)])
	}

	return sampled
}

func readIDs(ctx context.Context, conn *jump.Connector, params common.ReadParams) (datautils.StringSet, error) {
	params.Fields = connectors.Fields("id")

	rows, err := readAll(ctx, conn, params)
	if err != nil {
		return nil, err
	}

	ids := datautils.NewStringSet()
	for _, row := range rows {
		ids.AddOne(row.Id)
	}

	return ids, nil
}

// readAll follows NextPage until the connector reports the last page.
func readAll(ctx context.Context, conn *jump.Connector, params common.ReadParams) ([]common.ReadResultRow, error) {
	var rows []common.ReadResultRow

	for {
		res, err := conn.Read(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", params.ObjectName, err)
		}

		rows = append(rows, res.Data...)

		if res.Done {
			return rows, nil
		}

		params.NextPage = res.NextPage
	}
}
