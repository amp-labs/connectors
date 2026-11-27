package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// Reads every page of an object and reports the pages and distinct record ids, e.g.
//
//	go run ./test/twenty/read -object vendorContracts -page-size 2 -fields name,renewalNotes
//	go run ./test/twenty/read -object vendorContracts -since 2026-10-25T13:00:00.000Z -until 2026-10-25T13:05:00.000Z
//	go run ./test/twenty/read -object vendorContracts -deleted
func main() {
	objectName := flag.String("object", "companies", "object name (namePlural)")
	pageSize := flag.Int("page-size", 0, "page size, 0 for the maximum")
	fields := flag.String("fields", "id,updatedAt", "comma-separated fields")
	since := flag.String("since", "", "RFC 3339 lower bound on updatedAt")
	until := flag.String("until", "", "RFC 3339 upper bound on updatedAt")
	deleted := flag.Bool("deleted", false, "read soft-deleted records (the trash) instead")
	flag.Parse()

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetTwentyConnector(ctx)

	params := common.ReadParams{
		ObjectName: *objectName,
		Fields:     connectors.Fields(strings.Split(*fields, ",")...),
		PageSize:   *pageSize,
		Since:      parseTime(*since),
		Until:      parseTime(*until),
		Deleted:    *deleted,
	}

	ids := map[string]int{}
	pages := 0

	for {
		res, err := conn.Read(ctx, params)
		if err != nil {
			utils.Fail("error reading from Twenty", "object", *objectName, "error", err)
		}

		pages++

		for _, row := range res.Data {
			ids[row.Id]++
		}

		slog.Info("page", "number", pages, "rows", res.Rows, "nextPage", res.NextPage, "done", res.Done)
		utils.DumpJSON(res, os.Stdout)

		if res.Done {
			break
		}

		params.NextPage = res.NextPage
	}

	slog.Info("read complete", "object", *objectName, "pages", pages, "distinctIds", len(ids), "ids", ids)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		utils.Fail("invalid time", "value", value, "error", err)
	}

	return t
}
