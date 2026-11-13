package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// Live, read-only check of the jobs currentWorkflowStatus field: reads one
// page of jobs, then the same jobs by id, each with and without the field
// selected, and reports the API calls each made by endpoint. Without the
// field there must be no milestones/current calls; with it, one per job.
//
// Prints ids, timestamps and the status field only — no job names, addresses
// or contact data.

const pageSize = 5

// countingTransport tallies outgoing requests by endpoint kind.
type countingTransport struct {
	mu     sync.Mutex
	counts map[string]int
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.counts[endpointKind(req.URL.Path)]++
	t.mu.Unlock()

	return http.DefaultTransport.RoundTrip(req)
}

// take returns the counts since the last call and resets them.
func (t *countingTransport) take() map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := t.counts
	t.counts = map[string]int{}

	return out
}

func endpointKind(path string) string {
	switch parts := strings.Split(strings.TrimPrefix(path, "/api/v2/"), "/"); {
	case len(parts) == 1:
		return "GET /" + parts[0]
	case parts[0] == "jobs" && len(parts) == 2:
		return "GET /jobs/{id}"
	case parts[0] == "jobs" && len(parts) >= 3:
		return "GET /jobs/{id}/" + strings.Join(parts[2:], "/")
	default:
		return "GET /" + strings.Join(parts, "/")
	}
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	// The harness client uses http.DefaultClient; count everything it sends.
	counter := &countingTransport{counts: map[string]int{}}
	http.DefaultClient.Transport = counter

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	var jobIDs []string

	for _, fields := range [][]string{
		{"id", "modifiedDate"},
		{"id", "modifiedDate", "currentWorkflowStatus"},
	} {
		ids, err := readJobs(ctx, conn, fields)
		if err != nil {
			utils.Fail("read jobs", "error", err)
		}

		slog.Info("Read jobs API calls", "fields", fields, "calls", counter.take())

		jobIDs = ids
	}

	if len(jobIDs) > 3 {
		jobIDs = jobIDs[:3]
	}

	for _, fields := range [][]string{
		{"id", "modifiedDate"},
		{"id", "modifiedDate", "currentWorkflowStatus"},
	} {
		if err := readJobsByID(ctx, conn, jobIDs, fields); err != nil {
			utils.Fail("read jobs by id", "error", err)
		}

		slog.Info("GetRecordsByIds API calls", "fields", fields, "calls", counter.take())
	}
}

func readJobs(ctx context.Context, conn *acculynx.Connector, fields []string) ([]string, error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "jobs",
		Fields:     connectors.Fields(fields...),
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(res.Data))

	for _, row := range res.Data {
		ids = append(ids, row.Id)
		printRow("read", row)
	}

	return ids, nil
}

func readJobsByID(ctx context.Context, conn *acculynx.Connector, ids, fields []string) error {
	res, err := conn.GetRecordsByIds(ctx, "jobs", ids, fields, nil)
	if err != nil {
		return err
	}

	for _, row := range res.Rows {
		printRow("by-id", row)
	}

	return nil
}

// printRow prints the row's id, modifiedDate and currentWorkflowStatus. Read
// lowercases Fields keys while GetRecordsByIds keeps the requested spelling,
// so keys are looked up case-insensitively.
func printRow(source string, row common.ReadResultRow) {
	fields := common.ToStringMap(row.Fields)
	modified, _ := fields.GetCaseInsensitive("modifiedDate")

	value := "<absent>"
	if status, ok := fields.GetCaseInsensitive("currentWorkflowStatus"); ok {
		value = fmt.Sprintf("%v", status)
	}

	fmt.Printf("  %-5s job=%s modifiedDate=%v currentWorkflowStatus=%s\n", //nolint:forbidigo
		source, row.Id, modified, value)
}
