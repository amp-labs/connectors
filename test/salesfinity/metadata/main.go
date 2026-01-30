package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/salesfinity"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := salesfinity.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"analytics/list-performance",
		"analytics/sdr-performance",
		"call-log",
		"contact-lists/csv",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
