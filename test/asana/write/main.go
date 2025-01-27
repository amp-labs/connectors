package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/asana"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testWriteProjects(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testWriteProjects(ctx context.Context) error {
	conn := asana.GetAsanaConnector(ctx)

	params := common.WriteParams{
		ObjectName: "projects",
		RecordData: map[string]any{
			"data": map[string]any{
				"name":         "Stuff to buy",
				"archived":     false,
				"color":        "light-green",
				"default_view": "calendar",
				"due_date":     "2019-09-15",
				"due_on":       "2019-09-15",
				"team":         "1209100536982881",
				"workspace":    "1206661566061885",
			},
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error(err.Error())
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
