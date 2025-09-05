package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/brevo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := brevo.GetBrevoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"attributes/deals", "blockedContacts", "children", "companies/attributes", "emailCampaigns"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
