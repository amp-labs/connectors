package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	br "github.com/amp-labs/connectors/providers/bitbucket"
	"github.com/amp-labs/connectors/test/bitbucket"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()
	conn := bitbucket.GetConnector(ctx)

	if err := createProject(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateProject(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createSnippet(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	return 0
}

func createProject(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "projects",
		RecordData: map[string]any{
			"name":        "Kronk Project",
			"key":         "kronk",
			"description": "Software for building Talking Agents",
			"is_private":  false,
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateProject(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "projects",
		RecordId:   "MARS",
		RecordData: map[string]any{
			"description": "Software for colonizing planets.",
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createSnippet(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "hooks",
		RecordData: map[string]any{
			"description": "Webhook Description",
			"url":         "https://example.com/",
			"active":      true,
			"secret":      "this is a really bad secret",
			"events": []string{
				"repo:push",
				"issue:created",
				"issue:updated",
			},
		}}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
