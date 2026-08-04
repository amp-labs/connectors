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
	"github.com/amp-labs/connectors/providers/mailgun"
	testMailgun "github.com/amp-labs/connectors/test/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

const maxPages = 3

type paginationCase struct {
	name       string
	objectName string
	fields     []string
	pageSize   int
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testMailgun.GetMailgunConnector(ctx)

	cases := []paginationCase{
		{
			name:       "paging_next",
			objectName: "templates",
			fields:     []string{"id", "name"},
			pageSize:   1,
		},
		{
			name:       "total_count_skip",
			objectName: "domains",
			fields:     []string{"id", "name", "state"},
			pageSize:   1,
		},
		{
			name:       "total_skip",
			objectName: "users",
			fields:     []string{"id", "email", "name"},
			pageSize:   1,
		},
		{
			name:       "paging_next_capital",
			objectName: "dynamic_pools/history",
			fields:     []string{"id", "timestamp", "domain_name"},
			pageSize:   10,
		},
		{
			name:       "limit_only",
			objectName: "bounce-classification/stats",
			fields:     []string{"entity-id", "rule-id", "short-explanation"},
			pageSize:   10,
		},
		{
			name:       "none",
			objectName: "webhooks",
			fields:     []string{"webhook_id", "url"},
			pageSize:   1,
		},
	}

	for _, tc := range cases {
		if err := testPagination(ctx, conn, tc); err != nil {
			slog.Error("pagination test failed", "type", tc.name, "object", tc.objectName, "error", err)
		}
	}
}

func testPagination(ctx context.Context, conn *mailgun.Connector, tc paginationCase) error {
	slog.Info("Testing pagination", "type", tc.name, "object", tc.objectName, "pageSize", tc.pageSize)

	params := common.ReadParams{
		ObjectName: tc.objectName,
		Fields:     connectors.Fields(tc.fields...),
	}
	if tc.pageSize > 0 {
		params.PageSize = tc.pageSize
	}

	for page := 1; page <= maxPages; page++ {
		res, err := conn.Read(ctx, params)
		if err != nil {
			return fmt.Errorf("read page %d: %w", page, err)
		}

		slog.Info("Read page",
			"type", tc.name,
			"object", tc.objectName,
			"page", page,
			"rows", res.Rows,
			"done", res.Done,
			"nextPage", res.NextPage,
		)

		utils.DumpJSON(res, os.Stdout)

		if res.Done || res.NextPage == "" {
			break
		}

		params.NextPage = res.NextPage
	}

	return nil
}
