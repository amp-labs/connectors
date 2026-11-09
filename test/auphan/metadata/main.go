package main

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/amp-labs/connectors/test/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := auphan.GetAuphanConnector(ctx)

	res, err := conn.ListObjectMetadata(ctx, []string{"payments", "sales", "stores"})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	for name, objErr := range res.Errors {
		fmt.Printf("%s: ERROR %v\n", name, objErr)
	}

	for name, metadata := range res.Result {
		fields := make([]string, 0, len(metadata.Fields))
		for id := range metadata.Fields {
			fields = append(fields, id)
		}

		slices.Sort(fields)
		fmt.Printf("%s (%d fields): %v\n", name, len(fields), fields)
	}

	if len(res.Errors) > 0 {
		os.Exit(1)
	}
}
