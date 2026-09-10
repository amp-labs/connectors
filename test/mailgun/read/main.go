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
	"github.com/amp-labs/connectors/providers/mailgun"
	testMailgun "github.com/amp-labs/connectors/test/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

// liveTestObjects covers every supported read object so a single run exercises
// each code path: account-scoped single-shot, offset and cursor pagination,
// domain-scoped objects (require the workspace domain), the lists/members
// fan-out, and the POST body-token pagination of analytics/logs.
//
// Some objects are plan-gated on trial/sandbox accounts and are expected to
// error there: ip_pools + dynamic_pools/domains (400 feature disabled),
// thresholds/limits + thresholds/alerts/send (403 not available on plan).
//
// dkim/keys is slow and flaky server-side: the unfiltered GET /v1/dkim/keys
// performs an account-wide key scan that takes 4-20s regardless of request
// parameters (verified with raw curl at limit 3/10/50/100 and with no
// parameters at all), and Mailgun returns 500 "Internal Server Error" when
// the scan trips its ~20s internal timeout. Filtered requests
// (?signing_domain=...) answer in ~1s, confirming the cost is the scan
// itself. Nothing the connector sends influences this; retries succeed.
//
//nolint:gochecknoglobals
var liveTestObjects = []string{
	// Account-scoped, single-shot.
	"webhooks",
	"keys",
	"ip_pools",
	"ips",
	"ip_whitelist",
	"accounts/subaccounts/ip_pools/all",
	"thresholds/limits",
	"thresholds/alerts/send",

	// Account-scoped, offset paginated.
	"domains",
	"routes",
	"lists",
	"accounts/subaccounts",
	"users",

	// Account-scoped, cursor paginated.
	"forwards",
	"dkim/keys",
	"account/templates",
	"dynamic_pools/domains",

	// Domain-scoped (needs workspace).
	"bounces",
	"complaints",
	"unsubscribes",
	"whitelists",
	"templates",
	"domains/credentials",
	"domains/keys",

	// Nested fan-out over mailing lists.
	"lists/members",

	// POST-sourced reads: logs (body-token pagination) and the modern Tags API
	// (body-carried skip/limit pagination).
	"analytics/logs",
	"analytics/tags",
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testMailgun.GetMailgunConnector(ctx)

	slog.Info("Live read test", "count", len(liveTestObjects))

	for _, obj := range liveTestObjects {
		slog.Info("Testing basic read", "object", obj)

		if err := testRead(ctx, conn, obj); err != nil {
			slog.Error(err.Error())
		}
	}

	// Pagination — domains is offset paginated and typically populated.
	slog.Info("Testing pagination for domains")

	if err := testPagination(ctx, conn, "domains"); err != nil {
		slog.Error(err.Error())
	}

	// Incremental read — Mailgun list endpoints take no time parameters, so
	// Since/Until sieve records connector-side by their timestamp field.
	slog.Info("Testing incremental read for domains")

	if err := testIncremental(ctx, conn, "domains"); err != nil {
		slog.Error(err.Error())
	}
}

// testIncremental reads the object with a wide Since (records included) and
// again with Since = now (records filtered out).
func testIncremental(ctx context.Context, conn *mailgun.Connector, objectName string) error {
	wide, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "name"),
		Since:      time.Now().AddDate(-10, 0, 0),
	})
	if err != nil {
		return fmt.Errorf("error reading %s with wide Since: %w", objectName, err)
	}

	slog.Info("Wide Since (10y)", "object", objectName, "rows", wide.Rows)

	narrow, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "name"),
		Since:      time.Now(),
	})
	if err != nil {
		return fmt.Errorf("error reading %s with narrow Since: %w", objectName, err)
	}

	slog.Info("Narrow Since (now)", "object", objectName, "rows", narrow.Rows)

	return nil
}

func testRead(ctx context.Context, conn *mailgun.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	slog.Info("Read result", "object", objectName, "rows", res.Rows, "nextPage", res.NextPage, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testPagination(ctx context.Context, conn *mailgun.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "name"),
		PageSize:   1,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 1: %w", objectName, err)
	}

	slog.Info("Page 1", "object", objectName, "rows", res.Rows, "nextPage", res.NextPage)
	utils.DumpJSON(res, os.Stdout)

	if res.NextPage == "" {
		slog.Warn("No next page found — pagination test may be incomplete", "object", objectName)

		return nil
	}

	params.NextPage = res.NextPage

	res2, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 2: %w", objectName, err)
	}

	slog.Info("Page 2", "object", objectName, "rows", res2.Rows, "done", res2.Done)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}
