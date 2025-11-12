package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/linkedin"
	"github.com/amp-labs/connectors/test/linkedin"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := linkedin.GetAdsConnector(ctx)

	err := testRead(context.Background(), conn, "adTargetingFacets", []string{"facetName", "entityTypes"}, time.Time{}, time.Time{})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "dmpEngagementSourceTypes", []string{""}, time.Time{}, time.Time{})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "adAccounts", []string{"id", "name"}, time.Time{}, time.Time{})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "adAnalytics", []string{""}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local), time.Time{})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "dmpSegments", []string{"name", "account"}, time.Time{}, time.Time{})
	if err != nil {
		return 1
	}

	return 0
}

func testRead(ctx context.Context, conn *ap.Connector, objName string, fields []string, since, until time.Time) error {
	params := common.ReadParams{
		ObjectName: objName,
		Fields:     connectors.Fields(fields...),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objName, err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
