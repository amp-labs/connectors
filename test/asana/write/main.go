package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/asana"
	"github.com/amp-labs/connectors/test/utils"
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

	utils.SetupLogging()

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
		utils.Fail("error writing to Asana", "error", err)
	}

	// Dump the result.
	utils.DumpJSON(res, os.Stdout)

	return nil
}
