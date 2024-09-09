package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/test/apollo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := apollo.GetApolloConnector(ctx, "apollo-creds.json")

	err := testReadOpportunitiesSearch(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadCustomFields(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadEmailAccounts(ctx, conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadOpportunitiesSearch(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "opportunities",
		Fields:     []string{"id"},
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadEmailAccounts(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "email_accounts",
		Fields:     []string{"user_id", "id", "email"},
		Since:      time.Now().Add(-1800 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadCustomFields(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "typed_custom_fields",
		Fields:     []string{"type", "id", "modality"},
		Since:      time.Now().Add(-1800 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
