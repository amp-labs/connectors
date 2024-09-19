package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amp-labs/connectors"
	"log"
	"os"

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

	err := testReadContactsSearch(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadPeopleSearch(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadOpportunitiesSearch(ctx, conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadContactsSearch(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id"),
		// NextPage:   "2",
	}

	res, err := conn.Search(ctx, params)
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

func testReadOpportunitiesSearch(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "opportunities",
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Search(ctx, params)
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

func testReadPeopleSearch(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "mixed_people",
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Search(ctx, params)
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
