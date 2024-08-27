package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/marketo"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	err := testWrite(context.Background())
	if err != nil {
		return 1
	}

	return 0
}

func testWrite(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorW(ctx)

	params := common.WriteParams{
		ObjectName: "leads",
		RecordData: map[string]any{
			"input": []map[string]any{
				{
					"email":     gofakeit.Email(),
					"firstName": gofakeit.Name(),
				},
			},
			"action":      "createOnly",
			"lookupField": "email",
		},
	}

	res, err := conn.Write(ctx, params)
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
