package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	gl "github.com/amp-labs/connectors/providers/gitlab"
	"github.com/amp-labs/connectors/test/gitlab"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := gitlab.GetConnector(ctx)

	err := testCreatingProjects(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingSnippets(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingProjects(ctx context.Context, conn *gl.Connector) error {
	params := common.WriteParams{
		ObjectName: "projects",
		RecordData: map[string]any{
			"name": "Test Project AB",
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

func testCreatingSnippets(ctx context.Context, conn *gl.Connector) error {
	params := common.WriteParams{
		ObjectName: "snippets",
		RecordData: map[string]any{
			"title":       "This is a snippet A",
			"description": "Hello World snippet A",
			"visibility":  "public",
			"files": []map[string]string{
				{
					"content":   "Hello world",
					"file_path": "testA.txt",
				},
			},
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
