package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	testConn "github.com/amp-labs/connectors/providers/kit"
	"github.com/amp-labs/connectors/test/kit"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := kit.GetKitConnector(context.Background())

	err := testReadCustomFields(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadTags(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadEmailTemplates(context.Background(), conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadCustomFields(ctx context.Context, conn *testConn.Connector) error {
	params := common.ReadParams{
		ObjectName: "custom_fields",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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

func testReadTags(ctx context.Context, conn *testConn.Connector) error {
	params := common.ReadParams{
		ObjectName: "tags",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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

func testReadEmailTemplates(ctx context.Context, conn *testConn.Connector) error {
	params := common.ReadParams{
		ObjectName: "email_templates",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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
