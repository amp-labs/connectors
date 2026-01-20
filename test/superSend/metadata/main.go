package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/superSend"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := superSend.GetSuperSendConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"teams",
		"senders",
		"sender-profiles",
		"labels",
		"contact/all",
		"campaigns/overview",
		"org",
		"managed-domains",
		"managed-mailboxes",
		"placement-tests",
		"auto/identitys",
		"conversation/latest-by-profile",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
