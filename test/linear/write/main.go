package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ln "github.com/amp-labs/connectors/providers/linear"
	"github.com/amp-labs/connectors/test/linear"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := linear.GetLinearConnector(ctx)

	err := testCreateIssue(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateDocument(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateProject(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreateIssue(ctx context.Context, conn *ln.Connector) error {
	params := common.WriteParams{
		ObjectName: "issues",
		RecordData: map[string]any{
			"title":  "this is a test issue",
			"teamId": "f34f6ac8-918c-4d7b-b1c3-25e8e53b8c0d",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
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

func testCreateDocument(ctx context.Context, conn *ln.Connector) error {
	params := common.WriteParams{
		ObjectName: "documents",
		RecordData: map[string]any{
			"title":   "this is a test document",
			"teamId":  "f34f6ac8-918c-4d7b-b1c3-25e8e53b8c0d",
			"content": "This is the content of the document.",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
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

func testCreateProject(ctx context.Context, conn *ln.Connector) error {
	params := common.WriteParams{
		ObjectName: "projects",
		RecordData: map[string]any{
			"name":    "this is a test project",
			"teamIds": []string{"f34f6ac8-918c-4d7b-b1c3-25e8e53b8c0d"},
			"content": "the content is empty",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
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
