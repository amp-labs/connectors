package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/test/attio"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := attio.GetAttioConnector(context.Background())

	err := testReadLists(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadWorkspacemembers(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadTasks(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadNotes(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadCompanies(context.Background(), conn)
	if err != nil {
		return 1
	}

	return 0
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

	// Print the results.
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

	// Print the results.
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

	// Print the results.
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

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadCompanies(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "companies",
		Fields:     connectors.Fields("name"),
		Since:      time.Date(2025, 3, 18, 0, 0, 0, 0, time.UTC),
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
