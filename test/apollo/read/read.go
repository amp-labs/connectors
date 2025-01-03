package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/test/apollo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := apollo.GetApolloConnector(ctx)

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

	err = testReadEmailerCampaigns(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadContacts(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadSequences(ctx, conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadOpportunitiesSearch(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "opportunities",
		Fields:     connectors.Fields("id"),
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
		Fields:     connectors.Fields("user_id", "id", "email"),
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
		Fields:     connectors.Fields("type", "id", "modality"),
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

func testReadEmailerCampaigns(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "emailer_campaigns",
		Fields:     connectors.Fields("id", "name", "archived"),
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

func testReadContacts(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "first_name", "name"),
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

func testReadSequences(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "sequences",
		Fields:     connectors.Fields("id", "name"),
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
