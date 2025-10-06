package main

import (
	"context"
	"os"

	meta "github.com/amp-labs/connectors/test/meta"

	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := meta.GetFacebookConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"users",
		"advertisable_applications",
		"minimum_budgets",
		"targetingbrowse",
		"customaudiencestos",
		"business_users",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
