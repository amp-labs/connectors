// nolint
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/test/attio"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := attio.GetAttioConnector(ctx)

	err := testReadObjects(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadLists(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadWorkspacemembers(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadWebhooks(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadTasks(ctx, conn)
	if err != nil {
		return 1
	}

	err = testReadNotes(ctx, conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadObjects(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "objects",
		Fields:     connectors.Fields(""),
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

func testReadLists(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "lists",
		Fields:     connectors.Fields(""),
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

func testReadWorkspacemembers(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "workspace_members",
		Fields:     connectors.Fields(""),
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

func testReadWebhooks(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "webhooks",
		Fields:     connectors.Fields(""),
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

func testReadTasks(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "tasks",
		Fields:     connectors.Fields(""),
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

func testReadNotes(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "notes",
		Fields:     connectors.Fields(""),
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
