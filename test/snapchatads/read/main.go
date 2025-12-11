package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/snapchatads"
	"github.com/amp-labs/connectors/test/snapchatads"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := snapchatads.GetConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	err = testRead(context.Background(), conn, "billingcenters", []string{"postal_code"})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "adaccounts", []string{"client_paying_invoices"})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "members", []string{"display_name"})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "roles", []string{"container_kind"})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "targeting/demographics/advanced_demographics", []string{"id", "name"})
	if err != nil {
		return 1
	}

	return 0
}

func testRead(ctx context.Context, conn *ap.Connector, objName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objName,
		Fields:     connectors.Fields(fields...),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objName, err)
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
