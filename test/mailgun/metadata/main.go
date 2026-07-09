package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := mailgun.GetMailgunConnector(ctx)
	m, err := conn.ListObjectMetadata(ctx, []string{
		"accounts/subaccounts",
		"domains",
		"thresholds/limits",
		"users",
		"webhooks",
	})
	if err != nil {
		log.Fatal(err)
	}
	utils.DumpJSON(m, os.Stdout)
}
