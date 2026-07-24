package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// Live validation for the appointment associations work:
//
//  1. Read calendars/appointments with AssociatedObjects=[jobs,users] and show
//     each appointment carries a "jobs" edge (jobId, when present) and a "users"
//     edge (the calendarId fan-out parent).
//  2. GetRecordsByIds("users", [realUser, companyCalendar]) proves skip-on-404:
//     the real user hydrates (with role), the company calendar id — which is not
//     a user and returns 404 — is silently omitted, and the call does NOT error.
//  3. GetRecordsByIds("jobs", [realJob]) confirms job hydration still works.
//
// IDs below are from the Hatch test tenant (see acculynx-creds.json); adjust if
// the tenant data changes.
const (
	realUserID        = "794f06f5-b10e-4ceb-88e2-d597b03c301a" // "Diane Hatch" — a person calendar == user
	companyCalendarID = "faba30e4-10bc-4c10-a266-29a45d572ca4" // "Hatch Homes Services" — a calendar that is NOT a user (404)
	realJobID         = "6c0b61d3-ca9e-4593-91bf-fa4df85485e6" // a jobId taken off a live appointment
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("=== 1) Read calendars/appointments with job+user associations ===")

	if err := readAppointmentsWithAssociations(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("=== 2) GetRecordsByIds(users) skips a non-user calendar id (404) ===")

	if err := hydrateUsersWithMissing(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("=== 3) GetRecordsByIds(jobs) hydrates a real job ===")

	if err := hydrateJob(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readAppointmentsWithAssociations(ctx context.Context, conn *acculynx.Connector) error {
	params := common.ReadParams{
		ObjectName:        "calendars/appointments",
		Fields:            connectors.Fields("id", "title", "jobId", "jobName", "eventType"),
		AssociatedObjects: []string{"jobs", "users"},
		Since:             time.Now().AddDate(0, 0, -60),
		Until:             time.Now(),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("reading appointments with associations: %w", err)
	}

	slog.Info("appointments read", "rows", res.Rows)

	withJob, withUser := 0, 0

	for _, row := range res.Data {
		if len(row.Associations["jobs"]) > 0 {
			withJob++
		}

		if len(row.Associations["users"]) > 0 {
			withUser++
		}
	}

	slog.Info("association coverage",
		"rows", res.Rows,
		"withJobEdge", withJob,
		"withUserEdge", withUser)

	// Dump the first few rows so the Associations map is visible.
	max := 3
	if len(res.Data) < max {
		max = len(res.Data)
	}

	utils.DumpJSON(res.Data[:max], os.Stdout)

	return nil
}

func hydrateUsersWithMissing(ctx context.Context, conn *acculynx.Connector) error {
	ids := []string{realUserID, companyCalendarID}

	rows, err := conn.GetRecordsByIds(ctx, "users", ids,
		[]string{"id", "displayName", "role"}, nil)
	if err != nil {
		return fmt.Errorf("GetRecordsByIds(users) errored — skip-on-404 is broken: %w", err)
	}

	slog.Info("users hydrated",
		"requested", len(ids),
		"returned", len(rows),
		"expected", "1 (company calendar id skipped)")

	utils.DumpJSON(rows, os.Stdout)

	return nil
}

func hydrateJob(ctx context.Context, conn *acculynx.Connector) error {
	rows, err := conn.GetRecordsByIds(ctx, "jobs", []string{realJobID},
		[]string{"id", "jobName"}, nil)
	if err != nil {
		return fmt.Errorf("GetRecordsByIds(jobs): %w", err)
	}

	slog.Info("job hydrated", "returned", len(rows))
	utils.DumpJSON(rows, os.Stdout)

	return nil
}
