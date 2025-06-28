package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/campaignmonitor"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := campaignmonitor.GetCampaignMonitorConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"clients", "admins", "campaigns"})
	if err != nil {
		utils.Fail("error listing metadata for campaign monitor", "error", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
