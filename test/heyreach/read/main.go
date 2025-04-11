package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/test/heyreach"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := heyreach.GetHeyreachConnector(context.Background())

	err := testReadCampaign(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadLIAccount(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadList(context.Background(), conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadCampaign(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "campaign/GetAll",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read campaign: %w", err)
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadLIAccount(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "li_account/GetAll",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read li_account: %w", err)
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadList(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "list/GetAll",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read li_account: %w", err)
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
