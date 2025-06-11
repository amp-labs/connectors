package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/facebook"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := facebook.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"customaudiences",
		"saved_audiences",
		"users",
		"account_controls",
		"advertisable_applications",
		"minimum_budgets",
		"targetingbrowse",
		"adimages",
		"customaudiencestos",
		"adrules_library",
		"adspixels",
		"business_users",
		"system_users",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
